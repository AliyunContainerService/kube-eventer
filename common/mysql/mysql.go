// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"k8s.io/klog"
	"net/url"
	"strings"
)

const (
	DEFAULT_TABLE = "k8s-event"
)

type MysqlService struct {
	Db      *sql.DB
	Table   string
	Dsn     string
	Cluster string
}

type MysqlKubeEventPoint struct {
	Cluster                  string
	Namespace                string
	Kind                     string
	Name                     string
	Type                     string
	Reason                   string
	Message                  string
	EventID                  string
	Source                   string
	FirstOccurrenceTimestamp string
	LastOccurrenceTimestamp  string
}

func (mySvc MysqlService) SaveData(sinkData []interface{}) error {

	if len(sinkData) == 0 {
		klog.Warningf("insert data is []")
		return nil
	}

	prepareStatement := fmt.Sprintf("INSERT INTO %s (cluster,namespace,kind,name,type,reason,message,event_id,source,first_occurrence_time,last_occurrence_time) VALUES(?,?,?,?,?,?,?,?,?,?,?)", mySvc.Table)

	// Prepare statement for inserting data
	stmtIns, err := mySvc.Db.Prepare(prepareStatement)
	if err != nil {
		klog.Errorf("failed to Prepare statement for inserting data ")
		return err
	}

	defer stmtIns.Close()

	for _, data := range sinkData {

		ked := data.(MysqlKubeEventPoint)
		klog.Infof("Begin Insert Mysql Data ...")
		klog.Infof("Cluster: %s, Namespace: %s, Kind: %s, Name: %s, Type: %s, Reason: %s, Message: %s, EventID: %s, Source: %s, FirstOccurrenceTimestamp: %s, LastOccurrenceTimestamp: %s ", ked.Cluster, ked.Namespace, ked.Kind, ked.Name, ked.Type, ked.Reason, ked.Message, ked.EventID, ked.Source, ked.FirstOccurrenceTimestamp, ked.LastOccurrenceTimestamp)
		_, err = stmtIns.Exec(ked.Cluster, ked.Namespace, ked.Kind, ked.Name, ked.Type, ked.Reason, ked.Message, ked.EventID, ked.Source, ked.FirstOccurrenceTimestamp, ked.LastOccurrenceTimestamp)
		if err != nil {
			klog.Errorf("failed to Prepare statement for inserting data ")
			return err
		}
		klog.Infof("Insert Mysql Data Suc...")

	}

	return nil
}

func (mySvc MysqlService) FlushData() error {
	return nil
}

func (mySvc MysqlService) CreateDatabase(name string) error {
	return nil
}

func (mySvc MysqlService) CloseDB() error {
	return mySvc.Db.Close()
}

func NewMysqlClient(uri *url.URL) (*MysqlService, error) {
	mysqlSvc := &MysqlService{}

	if uri.Query().Get("table") != "" {
		mysqlSvc.Table = uri.Query().Get("table")
		mysqlSvc.Cluster = uri.Query().Get("cluster")
		slice := strings.Split(uri.RawQuery, "&")
		mysqlSvc.Dsn = slice[0]
	} else {
		mysqlSvc.Table = DEFAULT_TABLE
		mysqlSvc.Dsn = uri.RawQuery
	}

	klog.Infof("mysql jdbc url: %s", mysqlSvc.Dsn)

	db, err := sql.Open("mysql", mysqlSvc.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect mysql according jdbc url string: %s", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot open a connection for mysql according jdbc url string: %s", err)
	}

	mysqlSvc.Db = db

	return mysqlSvc, nil
}
