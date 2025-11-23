package repo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

var (
	ErrNotFound    = errors.New("not found")
	ErrPRExists    = errors.New("pr exists")
	ErrPRMerged    = errors.New("pr merged")
	ErrNotAssigned = errors.New("not assigned")
	ErrNoCandidate = errors.New("no candidate")
)

type Store struct {
	pool *pgxpool.Pool
	rnd  *rand.Rand
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (s *Store) CreateTeam(ctx context.Context, t Team) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO teams(team_name) VALUES($1) ON CONFLICT DO NOTHING`,
		t.TeamName); err != nil {
		return err
	}

	for _, m := range t.Members {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO users(user_id, username, team_name, is_active) VALUES($1,$2,$3,$4)
			 ON CONFLICT (user_id) DO UPDATE SET username=EXCLUDED.username, team_name=EXCLUDED.team_name, is_active=EXCLUDED.is_active`,
			m.UserID, m.Username, t.TeamName, m.IsActive); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) GetTeam(ctx context.Context, teamName string) (Team, error) {
	rows, err := s.pool.Query(ctx, `SELECT user_id, username, is_active FROM users WHERE team_name=$1 ORDER BY user_id`, teamName)
	if err != nil {
		return Team{}, err
	}
	defer rows.Close()

	members := make([]TeamMember, 0)
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return Team{}, err
		}
		members = append(members, m)
	}
	if len(members) == 0 {
		return Team{}, ErrNotFound
	}
	return Team{TeamName: teamName, Members: members}, nil
}

func (s *Store) SetUserActive(ctx context.Context, userID string, isActive bool) (User, error) {
	cmd, err := s.pool.Exec(ctx, `UPDATE users SET is_active=$1 WHERE user_id=$2`, isActive, userID)
	if err != nil {
		return User{}, err
	}
	if cmd.RowsAffected() == 0 {
		return User{}, ErrNotFound
	}
	var u User
	if err := s.pool.QueryRow(ctx, `SELECT user_id, username, team_name, is_active FROM users WHERE user_id=$1`, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) CreatePR(ctx context.Context, prID, prName, authorID string) (PullRequest, error) {
	// start tx
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PullRequest{}, err
	}
	defer tx.Rollback(ctx)

	var teamName string
	if err := tx.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id=$1`, authorID).Scan(&teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PullRequest{}, ErrNotFound
		}
		return PullRequest{}, err
	}

	var existing string
	if err := tx.QueryRow(ctx, `SELECT pull_request_id FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&existing); err == nil {
		return PullRequest{}, ErrPRExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PullRequest{}, err
	}

	rows, err := tx.Query(ctx, `SELECT user_id FROM users WHERE team_name=$1 AND is_active=true AND user_id<>$2`, teamName, authorID)
	if err != nil {
		return PullRequest{}, err
	}
	defer rows.Close()

	candidates := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return PullRequest{}, err
		}
		candidates = append(candidates, id)
	}

	s.rnd.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	assign := candidates
	if len(assign) > 2 {
		assign = assign[:2]
	}

	if _, err := tx.Exec(ctx, `INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id, status, created_at) VALUES($1,$2,$3,'OPEN',now())`, prID, prName, authorID); err != nil {
		return PullRequest{}, err
	}

	for _, a := range assign {
		if _, err := tx.Exec(ctx, `INSERT INTO pr_reviewers(pull_request_id, reviewer_id) VALUES($1,$2)`, prID, a); err != nil {
			return PullRequest{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return PullRequest{}, err
	}

	pr := PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: assign,
	}
	return pr, nil
}

func (s *Store) MergePR(ctx context.Context, prID string) (PullRequest, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PullRequest{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PullRequest{}, ErrNotFound
		}
		return PullRequest{}, err
	}

	if status == "MERGED" {
		var pr PullRequest
		if err := tx.QueryRow(ctx, `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return PullRequest{}, err
		}
		rows, _ := tx.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id=$1`, prID)
		defer rows.Close()
		pr.AssignedReviewers = []string{}
		for rows.Next() {
			var r string
			_ = rows.Scan(&r)
			pr.AssignedReviewers = append(pr.AssignedReviewers, r)
		}
		return pr, nil
	}

	if _, err := tx.Exec(ctx, `UPDATE pull_requests SET status='MERGED', merged_at=now() WHERE pull_request_id=$1`, prID); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := tx.QueryRow(ctx, `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		return PullRequest{}, err
	}
	rows, _ := tx.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id=$1`, prID)
	defer rows.Close()
	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var r string
		_ = rows.Scan(&r)
		pr.AssignedReviewers = append(pr.AssignedReviewers, r)
	}

	if err := tx.Commit(ctx); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

func (s *Store) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if status == "MERGED" {
		return "", ErrPRMerged
	}

	var assigned bool
	if err := tx.QueryRow(ctx, `SELECT true FROM pr_reviewers WHERE pull_request_id=$1 AND reviewer_id=$2`, prID, oldReviewerID).Scan(&assigned); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotAssigned
		}
		return "", err
	}

	var team string
	if err := tx.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id=$1`, oldReviewerID).Scan(&team); err != nil {
		return "", err
	}

	var author string
	if err := tx.QueryRow(ctx, `SELECT author_id FROM pull_requests WHERE pull_request_id=$1`, prID).Scan(&author); err != nil {
		return "", err
	}

	rows, err := tx.Query(ctx, `SELECT user_id FROM users WHERE team_name=$1 AND is_active=true AND user_id<>$2 AND user_id<>$3 AND user_id NOT IN (SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id=$4)`, team, oldReviewerID, author, prID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	candidates := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return "", ErrNoCandidate
	}

	newID := candidates[s.rnd.Intn(len(candidates))]

	if _, err := tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pull_request_id=$1 AND reviewer_id=$2`, prID, oldReviewerID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO pr_reviewers(pull_request_id, reviewer_id) VALUES($1,$2)`, prID, newID); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return newID, nil
}

func (s *Store) GetPRsForReviewer(ctx context.Context, userID string) ([]PullRequestShort, error) {
	rows, err := s.pool.Query(ctx, `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status FROM pull_requests pr JOIN pr_reviewers r ON r.pull_request_id = pr.pull_request_id WHERE r.reviewer_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []PullRequestShort{}
	for rows.Next() {
		var p PullRequestShort
		if err := rows.Scan(&p.PullRequestID, &p.PullRequestName, &p.AuthorID, &p.Status); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

type ReviewerStat struct {
	UserID string `json:"user_id"`
	Count  int64  `json:"review_count"`
}

func (s *Store) GetReviewerAssignmentStats(ctx context.Context) ([]ReviewerStat, error) {
	rows, err := s.pool.Query(ctx, `
SELECT reviewer_id, COUNT(*) as cnt
FROM pr_reviewers
GROUP BY reviewer_id
ORDER BY cnt DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]ReviewerStat, 0)
	for rows.Next() {
		var r ReviewerStat
		if err := rows.Scan(&r.UserID, &r.Count); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

type BulkDeactivateResult struct {
	TeamName           string            `json:"team_name"`
	DeactivatedUsers   []string          `json:"deactivated_users"`
	ReassignedPRsCount int               `json:"reassigned_prs_count"`
	ReassignFailures   map[string]string `json:"reassign_failures"` // prID -> error
}

func (s *Store) BulkDeactivateTeam(ctx context.Context, teamName string) (BulkDeactivateResult, error) {
	res := BulkDeactivateResult{TeamName: teamName, ReassignFailures: map[string]string{}}

	// 1) Получить список пользователей команды
	rows, err := s.pool.Query(ctx, `SELECT user_id FROM users WHERE team_name=$1`, teamName)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	users := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return res, err
		}
		users = append(users, id)
	}

	if len(users) == 0 {
		return res, ErrNotFound
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE users SET is_active = false WHERE team_name=$1`, teamName); err != nil {
		return res, err
	}

	prRows, err := tx.Query(ctx, `
SELECT DISTINCT pr.pull_request_id
FROM pull_requests pr
JOIN pr_reviewers r ON r.pull_request_id = pr.pull_request_id
JOIN users u ON u.user_id = r.reviewer_id
WHERE pr.status = 'OPEN' AND u.team_name = $1
`, teamName)
	if err != nil {
		return res, err
	}
	defer prRows.Close()

	prIDs := []string{}
	for prRows.Next() {
		var id string
		if err := prRows.Scan(&id); err != nil {
			return res, err
		}
		prIDs = append(prIDs, id)
	}
	reassigned := 0
	for _, prID := range prIDs {
		rRows, err := tx.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id=$1`, prID)
		if err != nil {
			res.ReassignFailures[prID] = err.Error()
			continue
		}
		cur := []string{}
		for rRows.Next() {
			var rid string
			_ = rRows.Scan(&rid)
			cur = append(cur, rid)
		}
		rRows.Close()
		for _, old := range cur {
			var team string
			var isActive bool
			if err := tx.QueryRow(ctx, `SELECT team_name, is_active FROM users WHERE user_id=$1`, old).Scan(&team, &isActive); err != nil {
				res.ReassignFailures[prID] = fmt.Sprintf("check user: %v", err)
				continue
			}
			if team != teamName || isActive {
				continue
			}
			reassigned++
		}
	}
	res.DeactivatedUsers = users
	res.ReassignedPRsCount = reassigned

	if err := tx.Commit(ctx); err != nil {
		return res, err
	}

	return res, nil
}
