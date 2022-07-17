package integration

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/vehsamrak/gorep"
)

var testDatabase *sqlx.DB

func TestMain(m *testing.M) {
	const (
		maxDockerWaitSeconds = 120
	)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "11",
			Env: []string{
				"POSTGRES_PASSWORD=secret",
				"POSTGRES_USER=user_name",
				"POSTGRES_DB=dbname",
				"listen_addresses = '*'",
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		panic(fmt.Errorf("[docker_test] could not start resource: %w", err))
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	err = resource.Expire(maxDockerWaitSeconds)
	if err != nil {
		panic(fmt.Errorf("[docker_test] expiration error: %w", err))
	}

	pool.MaxWait = maxDockerWaitSeconds * time.Second
	if err = pool.Retry(
		func() error {
			testDatabase, err = sqlx.Connect("postgres", databaseUrl)

			return err
		},
	); err != nil {
		panic(fmt.Errorf("[create_database] error: %w", err))
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("[docker_test] could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTable(database *sqlx.DB, tableName string, columnsMap map[string]string) {
	if len(columnsMap) == 0 {
		panic(fmt.Errorf("[create_table] no columns specified for create table operation"))
	}

	columns := make([]string, 0, len(columnsMap))
	for columnName, columnType := range columnsMap {
		columns = append(columns, fmt.Sprintf("%s %s", columnName, columnType))
	}

	query := fmt.Sprintf("create table %s (%s)", tableName, strings.Join(columns, ", "))
	_, err := database.Exec(query)
	if err != nil {
		panic(fmt.Errorf("[create_table] error: %w", err))
	}
}

func dropTable(database gorep.Database, tableName string) {
	_, err := database.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		panic(fmt.Errorf("[drop_table] error: %w", err))
	}
}
