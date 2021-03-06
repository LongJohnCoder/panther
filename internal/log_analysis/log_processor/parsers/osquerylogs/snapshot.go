package osquerylogs

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
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers/timestamp"
)

var SnapshotDesc = `Snapshot contains all the data included in OsQuery differential logs
Reference: https://osquery.readthedocs.io/en/stable/deployment/logging/`

type Snapshot struct {
	Action         *string                `json:"action,omitempty" validate:"required,eq=snapshot"`
	CalendarTime   *timestamp.ANSICwithTZ `json:"calendarTime,omitempty" validate:"required"`
	Counter        *int                   `json:"counter,omitempty,string" validate:"required"`
	Decorations    map[string]string      `json:"decorations,omitempty"`
	Epoch          *int                   `json:"epoch,omitempty,string" validate:"required"`
	HostIdentifier *string                `json:"hostIdentifier,omitempty" validate:"required"`
	Name           *string                `json:"name,omitempty" validate:"required"`
	Snapshot       []map[string]string    `json:"snapshot,omitempty" validate:"required"`
	UnixTime       *int                   `json:"unixTime,omitempty,string" validate:"required"`
}

// SnapshotParser parses OsQuery snapshot logs
type SnapshotParser struct{}

// Parse returns the parsed events or nil if parsing failed
func (p *SnapshotParser) Parse(log string) []interface{} {
	event := &Snapshot{}
	err := jsoniter.UnmarshalFromString(log, event)
	if err != nil {
		zap.L().Debug("failed to unmarshal log", zap.Error(err))
		return nil
	}

	if err := parsers.Validator.Struct(event); err != nil {
		zap.L().Debug("failed to validate log", zap.Error(err))
		return nil
	}
	return []interface{}{event}
}

// LogType returns the log type supported by this parser
func (p *SnapshotParser) LogType() string {
	return "Osquery.Snapshot"
}
