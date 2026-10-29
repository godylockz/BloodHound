// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
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
//
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package azure_test

import (
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	graphAzure "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleMembersHonourAssignmentScope(t *testing.T) {
	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		tenantID      = strings.ToUpper(integration.RandomObjectID(t))
		tenant        = NewAzureTenant(t, &suite, tenantID)
		role          = NewAzureRole(t, &suite, "HelpdeskAdminRole", integration.RandomObjectID(t), graphAzure.HelpdeskAdministratorRole, tenantID)
		tenantScoped  = NewAzureServicePrincipal(t, &suite, "TenantScoped", integration.RandomObjectID(t), tenantID)
		unscoped      = NewAzureServicePrincipal(t, &suite, "Unscoped", integration.RandomObjectID(t), tenantID)
		adminUnitOnly = NewAzureServicePrincipal(t, &suite, "AdminUnitScoped", integration.RandomObjectID(t), tenantID)
		appOnly       = NewAzureServicePrincipal(t, &suite, "AppScoped", integration.RandomObjectID(t), tenantID)
	)

	scope := func(value string) *graph.Properties {
		return graph.AsProperties(graph.PropertyMap{graphAzure.Scope: value})
	}

	NewRelationship(t, &suite, tenant, role, graphAzure.Contains)
	NewRelationship(t, &suite, tenantScoped, role, graphAzure.HasRole, scope(tenantID))
	NewRelationship(t, &suite, unscoped, role, graphAzure.HasRole)
	NewRelationship(t, &suite, adminUnitOnly, role, graphAzure.HasRole, scope("ADMINISTRATIVEUNITS/"+strings.ToUpper(integration.RandomObjectID(t))))
	NewRelationship(t, &suite, appOnly, role, graphAzure.HasRole, scope(strings.ToUpper(integration.RandomObjectID(t))))

	err := suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		members, err := azure.RoleMembers(tx, tenant, graphAzure.HelpdeskAdministratorRole)
		require.NoError(t, err)

		assert.ElementsMatch(t, []graph.ID{tenantScoped.ID, unscoped.ID}, members.IDs())
		return nil
	})
	require.NoError(t, err)
}
