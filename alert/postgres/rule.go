package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/alert"
	"github.com/mainflux/mainflux/graphql"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errListAll           = errors.New("failed to query all from database")
)

var _ alert.RuleRepository = (*ruleRepository)(nil)

type ruleRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewRuleRepository(db *sqlx.DB, log logger.Logger) alert.RuleRepository {
	return &ruleRepository{db: db, log: log}
}

func (rr ruleRepository) ListAll() ([]alert.Rule, error) {
	var rules []alert.Rule

	q := "SELECT id, name, title, event_type, contents, notice FROM alert"
	rows, err := rr.db.Queryx(q)
	if err != nil {
		return []alert.Rule{}, err
	}

	for rows.Next() {
		var dbrl dbRule
		if err := rows.StructScan(&dbrl); err != nil {
			rr.log.Error(fmt.Sprintf("Failed to read retrieved rules due to %s", err))
			return []alert.Rule{}, err
		}

		rl, err := toRule(dbrl)
		if err != nil {
			rr.log.Error(fmt.Sprintf("Failed to deserialize rule due to %s", err))
			return []alert.Rule{}, err
		}

		rules = append(rules, rl)
	}

	return rules, nil
}

func toRule(dbrl dbRule) (alert.Rule, error) {
	rl := alert.Rule{
		Id: dbrl.Id,
	}

	if dbrl.Name.Valid {
		rl.Name = dbrl.Name.String
	}
	if dbrl.Title.Valid {
		rl.Title = dbrl.Title.String
	}
	rl.EventType = dbrl.EventType

	if err := json.Unmarshal([]byte(dbrl.Contents), &rl.Contents); err != nil {
		return alert.Rule{}, errors.Wrap(errListAll, err)
	}
	if err := json.Unmarshal([]byte(dbrl.Notice), &rl.Notice); err != nil {
		return alert.Rule{}, errors.Wrap(errListAll, err)
	}

	return rl, nil
}

type dbRule struct {
	Id       		string         		`db:"id"`
	Name  			sql.NullString 		`db:"name"`
	Title      		sql.NullString 		`db:"title"`
	EventType 		graphql.EventType 	`db:"event_type"`
	Contents 		string         		`db:"contents"`
	Notice 			string         		`db:"notice"`
}
