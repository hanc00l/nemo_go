package runner

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"testing"
	"time"
)

func TestCronTaskJob_Run(t *testing.T) {
	c := cron.New()
	eid, err := c.AddJob("@every 1s", CronTaskJob{"dj"})
	if err != nil {
		fmt.Println(err)
	}
	jobEntries["dj"] = eid
	c.Start()

	for k, v := range jobEntries {
		t.Log(k, v)
	}
	time.Sleep(2 * time.Second)
	c.Remove(jobEntries["dj"])

	time.Sleep(5 * time.Second)
}
