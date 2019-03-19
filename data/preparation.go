package data

import (
	"database/sql"
	"time"
	"github.com/sirupsen/logrus"
	"fmt"
)

type PreparationRecord struct {
	Uid         string
	ProgramType string
	Process     string
	Port        string
	Status      string
	Error       string
	CreateTime  string
	UpdateTime  string
}

const preparationTableDDL = `CREATE TABLE IF NOT EXISTS preparation (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uid VARCHAR(32) UNIQUE,
	program_type       VARCHAR NOT NULL,
	process    VARCHAR,
	port       VARCHAR,
	status     VARCHAR,
    error 	   VARCHAR,
	create_time VARCHAR,
	update_time VARCHAR
)`

var preIndexDDL = []string{
	`CREATE INDEX pre_uid_uidx ON preparation (uid)`,
	`CREATE INDEX pre_status_idx ON preparation (uid)`,
	`CREATE INDEX pre_type_process_idx ON preparation (program_type, process)`,
}

var insertPreDML = `INSERT INTO
	preparation (uid, program_type, process, port, status, error, create_time, update_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

func (s *Source) checkAndInitPreTable() {
	exists, err := s.preparationTableExists()
	if err != nil {
		logrus.Fatalf(err.Error())
	}
	if !exists {
		err = s.initPreparationTable()
		if err != nil {
			logrus.Fatalf(err.Error())
		}
	}
}

func (s *Source) initPreparationTable() error {
	_, err := s.DB.Exec(preparationTableDDL)
	if err != nil {
		return fmt.Errorf("create preparation table err, %s", err)
	}
	for _, sql := range preIndexDDL {
		s.DB.Exec(sql)
	}
	return nil
}

func (s *Source) preparationTableExists() (bool, error) {
	stmt, err := s.DB.Prepare(tableExistsDQL)
	if err != nil {
		return false, fmt.Errorf("select preparation table exists err when invoke db prepare, %s", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query("preparation")
	if err != nil {
		return false, fmt.Errorf("select preparation table exists or not err, %s", err)
	}
	defer rows.Close()
	var c int
	for rows.Next() {
		rows.Scan(&c)
		break
	}
	return c != 0, nil
}

func (s *Source) InsertPreparationRecord(record *PreparationRecord) error {
	stmt, err := s.DB.Prepare(insertPreDML)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		record.Uid,
		record.ProgramType,
		record.Process,
		record.Port,
		record.Status,
		record.Error,
		record.CreateTime,
		record.UpdateTime,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Source) QueryPreparationByUid(uid string) (*PreparationRecord, error) {
	stmt, err := s.DB.Prepare(`SELECT * FROM preparation WHERE uid = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records, err := getPreparationRecordFrom(rows)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func (s *Source) QueryRunningPreByTypeAndProcess(programType string, process string) (*PreparationRecord, error) {
	query := `SELECT * FROM preparation WHERE program_type = ? and process = ? and status = "Running"`
	if process == "" {
		query = `SELECT * FROM preparation WHERE program_type = ? and status = "Running"`
	}
	stmt, err := s.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	if process == "" {
		rows, err = stmt.Query(programType)
	} else {
		rows, err = stmt.Query(programType, process)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records, err := getPreparationRecordFrom(rows)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func (s *Source) ListPreparationRecords() ([]*PreparationRecord, error) {
	stmt, err := s.DB.Prepare(`SELECT * FROM preparation`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	defer rows.Close()
	return getPreparationRecordFrom(rows)
}

func getPreparationRecordFrom(rows *sql.Rows) ([]*PreparationRecord, error) {
	records := make([]*PreparationRecord, 0)
	for rows.Next() {
		var id int
		var uid, t, p, port, status, error, createTime, updateTime string
		err := rows.Scan(&id, &uid, &t, &p, &port, &status, &error, &createTime, &updateTime)
		if err != nil {
			return nil, err
		}
		record := &PreparationRecord{
			Uid:         uid,
			ProgramType: t,
			Process:     p,
			Port:        port,
			Status:      status,
			Error:       error,
			CreateTime:  createTime,
			UpdateTime:  updateTime,
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Source) UpdatePreparationRecordByUid(uid, status, errMsg string) error {
	stmt, err := s.DB.Prepare(`UPDATE preparation
	SET status = ?, error = ?, update_time = ?
	WHERE uid = ?
`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(status, errMsg, time.Now().Format(time.RFC3339Nano), uid)
	if err != nil {
		return err
	}
	return nil
}