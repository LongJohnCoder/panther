package remediation

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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"

	"github.com/panther-labs/panther/api/gateway/remediation/models"
)

//InvokerAPI is the interface for the Invoker,
// the component that is responsible for invoking Remediation Lambda
type InvokerAPI interface {
	Remediate(*models.RemediateResource) error
	GetRemediations() (*models.Remediations, error)
}

//Invoker is responsible for invoking Remediation Lambda
type Invoker struct {
	lambdaClient lambdaiface.LambdaAPI
}

//NewInvoker method returns a new instance of Invoker
func NewInvoker(sess *session.Session) *Invoker {
	return &Invoker{
		lambdaClient: lambda.New(sess),
	}
}
