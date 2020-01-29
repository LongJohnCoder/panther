package nginx

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"encoding/csv"
	"strings"

	"go.uber.org/zap"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers/timestamp"
)

const (
	accessNumberOfColumns          = 9
	accessUserIdentifier           = "-"
	accessTimestampFormatTimeLocal = "06/Jan/2006:15:04:05 -0700"
)

var AccessDesc = `Access Logs for your Nginx server. We currently support 'combined' format. 
Reference: http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format`

//'$remote_addr - $remote_user [$time_local] '
//'"$request" $status $body_bytes_sent '
//'"$http_referer" "$http_user_agent"';
type Access struct {
	RemoteAddress *string            `json:"remoteAddr,omitempty"`
	RemoteUser    *string            `json:"remoteUser,omitempty"`
	Time          *timestamp.RFC3339 `json:"time" validate:"required"`
	Request       *string            `json:"request,omitempty"`
	Status        *int8              `json:"status,omitempty"`
	BodyBytesSent *int               `json:"bodyBytesSent,omitempty"`
	HttpReferer   *string            `json"httpReferer,omitempty"`
	HttpUserAgent *string            `json:"httpUserAgent,omitempty"`
}

// ALBParser parses AWS Application Load Balancer logs
type AccessParser struct{}

// Parse returns the parsed events or nil if parsing failed
func (p *AccessParser) Parse(log string) []interface{} {
	reader := csv.NewReader(strings.NewReader(log))
	// Separator between fields is the empty space
	reader.Comma = ' '

	records, err := reader.ReadAll()
	if len(records) == 0 || err != nil {
		zap.L().Debug("failed to parse the log as csv")
		return nil
	}

	// parser should only receive 1 line at a time
	if len(records) > 1 {
		zap.L().Debug("failed to parse the log as csv")
		return nil
	}
	record := records[0]

	if len(record) != accessNumberOfColumns {
		zap.L().Debug("failed to parse the log as csv (wrong number of columns)")
		return nil
	}

	if record[1] != accessUserIdentifier {
		zap.L().Debug("failed to parse the log as csv (user identifier should always be '-')")
		return nil
	}

	time, err := timestamp.Parse(accessTimestampFormatTimeLocal, record[3])
	if err != nil {
		zap.L().Debug("failed to parse time", zap.Error(err))
		return nil
	}

	event := &Access{
		RemoteAddress: parsers.CsvStringToPointer(record[0]),
		RemoteUser:    parsers.CsvStringToPointer(record[2]),
		Time:          &time,
		Request:       parsers.CsvStringToPointer(record[4]),
		Status:        parsers.CsvStringToInt8Pointer(record[5]),
		BodyBytesSent: parsers.CsvStringToIntPointer(record[6]),
		HttpReferer:   parsers.CsvStringToPointer(record[7]),
		HttpUserAgent: parsers.CsvStringToPointer(record[8]),
	}

	if err := parsers.Validator.Struct(event); err != nil {
		zap.L().Debug("failed to validate log", zap.Error(err))
		return nil
	}

	return []interface{}{event}
}

// LogType returns the log type supported by this parser
func (p *AccessParser) LogType() string {
	return "Nginx.Access"
}
