package api

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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/panther-labs/panther/pkg/genericapi"
)

//
type SqsPolicy struct {
	Version    string               `json:"Version"`
	Statements []SqsPolicyStatement `json:"Statement"`
}

type SqsPolicyStatement struct {
	SID       string            `json:"Sid"`
	Effect    string            `json:"Effect"`
	Principal map[string]string `json:"Principal"`
	Action    string            `json:"Action"`
	Resource  string            `json:"Resource"`
	Condition interface{}       `json:"Condition"`
}

const (
	sidFormat           = "PantherSubscriptionSID-%s"
	policyAttributeName = "Policy"
)

// AddPermissionToLogProcessorQueue modifies the SQS Queue policy of the Log Processor
// to allow SNS topic from new account to subscribe to it
func AddPermissionToLogProcessorQueue(accountID string) error {
	existingPolicy, err := getQueuePolicy()
	if err != nil {
		return err
	}
	if existingPolicy == nil {
		existingPolicy = &SqsPolicy{
			Version:    "2008-10-17",
			Statements: []SqsPolicyStatement{},
		}
	}

	if findStatementIndex(existingPolicy, accountID) >= 0 {
		errMsg := "account: " + accountID + " has already been configured"
		err = errors.WithStack(&genericapi.AlreadyExistsError{Message: errMsg})
		zap.L().Error(errMsg,
			zap.String("sqsQueueARN", logProcessorQueueArn),
			zap.String("awsAccountId", accountID),
			zap.Error(err))

		// // Returning user friendly message
		return err
	}

	existingPolicy.Statements = append(existingPolicy.Statements, getStatementForAccount(accountID))
	err = setQueuePolicy(existingPolicy)
	if err != nil {
		zap.L().Error("failed to set policy", zap.Error(errors.Wrap(err, "failed to set policy")))
	}
	return setQueuePolicy(existingPolicy)
}

// RemovePermissionFromLogProcessorQueue modifies the SQS Queue policy of the Log Processor
// so that SNS topics from that account cannot subscribe to the queue
func RemovePermissionFromLogProcessorQueue(accountID string) error {
	existingPolicy, err := getQueuePolicy()
	if err != nil {
		return err
	}
	if existingPolicy == nil {
		return errors.New("policy doesn't exist")
	}

	statementToRemoveIndex := findStatementIndex(existingPolicy, accountID)
	if statementToRemoveIndex < 0 {
		err := errors.New("didn't find expected statement in queue policy")
		zap.L().Error("didn't find expected statement in queue policy",
			zap.String("accountId", accountID),
			zap.Error(errors.Wrap(err, "didn't find expected statement in queue policy")))
		return err
	}
	// Remove statement
	existingPolicy.Statements[statementToRemoveIndex] = existingPolicy.Statements[len(existingPolicy.Statements)-1]
	existingPolicy.Statements = existingPolicy.Statements[:len(existingPolicy.Statements)-1]

	return setQueuePolicy(existingPolicy)
}

func getQueuePolicy() (*SqsPolicy, error) {
	getAttributesInput := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{policyAttributeName}),
		QueueUrl:       aws.String(logProcessorQueueURL),
	}
	attributes, err := SQSClient.GetQueueAttributes(getAttributesInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get queue attributes")
	}
	policyAttribute := attributes.Attributes[policyAttributeName]
	if policyAttribute == nil {
		return nil, nil
	}
	var policy SqsPolicy
	err = jsoniter.UnmarshalFromString(*policyAttribute, &policy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall queue policy")
	}
	return &policy, nil
}

func findStatementIndex(policy *SqsPolicy, accountID string) int {
	newStatementSid := fmt.Sprintf(sidFormat, accountID)
	for i, statement := range policy.Statements {
		if statement.SID == newStatementSid {
			return i
		}
	}
	return -1
}

func setQueuePolicy(policy *SqsPolicy) error {
	serializedPolicy, err := jsoniter.MarshalToString(policy)
	if err != nil {
		zap.L().Error("failed to serialize policy", zap.Error(errors.WithStack(err)))
		return errors.WithStack(err)
	}

	setAttributesInput := &sqs.SetQueueAttributesInput{
		Attributes: map[string]*string{
			policyAttributeName: aws.String(serializedPolicy),
		},
		QueueUrl: aws.String(logProcessorQueueURL),
	}
	_, err = SQSClient.SetQueueAttributes(setAttributesInput)
	if err != nil {
		return errors.Wrap(err, "failed to set queue attributes")
	}
	return nil
}

func getStatementForAccount(accountID string) SqsPolicyStatement {
	newStatementSid := fmt.Sprintf(sidFormat, accountID)
	return SqsPolicyStatement{
		SID:       newStatementSid,
		Effect:    "Allow",
		Principal: map[string]string{"AWS": "*"},
		Action:    "sqs:SendMessage",
		Resource:  logProcessorQueueArn,
		Condition: map[string]interface{}{
			"ArnLike": map[string]string{
				"aws:SourceArn": fmt.Sprintf("arn:aws:sns:*:%s:*", accountID),
			},
		},
	}
}
