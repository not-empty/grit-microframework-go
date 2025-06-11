package models

import "github.com/not-empty/grit-microframework-go/app/helper"

func init() {
	helper.RegisterRawQueries("example", map[string]string{
		"count": `
      SELECT
        COUNT(1) as total
      FROM example
    `,
	})
}
