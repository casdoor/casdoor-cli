// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import "strings"

// resource is a kind of Casdoor object. Casdoor's API is the same for all of them:
// get-<plural>?owner=, get-<singular>?id=owner/name, add-<singular>, update-<singular>?id=
// and delete-<singular>.
type resource struct {
	Plural   string
	Singular string
	Aliases  []string
	Short    string
	Columns  []string

	// Organizations, applications and tokens are owned by "admin", the others by an
	// organization.
	AdminOwned bool
	// The field that names the object, "name" unless set.
	NameField string
	// The query parameter of the id, "id" unless set.
	IdParam string
	// How many parts the id has, 2 ("owner/name") unless set.
	IdParts int
	// Only listing is supported.
	ListOnly bool
}

var defaultColumns = []string{"owner", "name", "displayName", "createdTime"}

var resources = []*resource{
	{Plural: "users", Singular: "user", Aliases: []string{"user"}, Short: "Users of an organization",
		Columns: []string{"owner", "name", "displayName", "email", "phone", "isAdmin", "isForbidden", "createdTime"}},
	{Plural: "groups", Singular: "group", Aliases: []string{"group"}, Short: "Groups of users",
		Columns: []string{"owner", "name", "displayName", "type", "parentId", "isEnabled"}},
	{Plural: "roles", Singular: "role", Aliases: []string{"role"}, Short: "Roles",
		Columns: []string{"owner", "name", "displayName", "users", "roles", "isEnabled"}},
	{Plural: "permissions", Singular: "permission", Aliases: []string{"permission", "perms"}, Short: "Permissions",
		Columns: []string{"owner", "name", "displayName", "users", "roles", "resources", "actions", "effect", "isEnabled"}},
	{Plural: "models", Singular: "model", Aliases: []string{"model"}, Short: "Casbin models"},
	{Plural: "adapters", Singular: "adapter", Aliases: []string{"adapter"}, Short: "Casbin adapters",
		Columns: []string{"owner", "name", "type", "databaseType", "host", "database", "table"}},
	{Plural: "enforcers", Singular: "enforcer", Aliases: []string{"enforcer"}, Short: "Casbin enforcers",
		Columns: []string{"owner", "name", "displayName", "model", "adapter"}},
	{Plural: "organizations", Singular: "organization", Aliases: []string{"organization", "orgs", "org"}, Short: "Organizations", AdminOwned: true,
		Columns: []string{"name", "displayName", "websiteUrl", "passwordType", "createdTime"}},
	{Plural: "applications", Singular: "application", Aliases: []string{"application", "apps", "app"}, Short: "Applications", AdminOwned: true,
		Columns: []string{"name", "displayName", "organization", "clientId", "createdTime"}},
	{Plural: "providers", Singular: "provider", Aliases: []string{"provider"}, Short: "Providers (OAuth, email, SMS, storage, payment...)",
		Columns: []string{"owner", "name", "displayName", "category", "type", "createdTime"}},
	{Plural: "certs", Singular: "cert", Aliases: []string{"cert"}, Short: "Certificates and keys",
		Columns: []string{"owner", "name", "displayName", "scope", "type", "cryptoAlgorithm", "expireInYears"}},
	{Plural: "tokens", Singular: "token", Aliases: []string{"token"}, Short: "OAuth tokens", AdminOwned: true,
		Columns: []string{"name", "application", "organization", "user", "expiresIn", "createdTime"}},
	{Plural: "sessions", Singular: "session", Aliases: []string{"session"}, Short: "Login sessions, the id is owner/name/application",
		Columns: []string{"owner", "name", "application", "sessionId", "createdTime"}, IdParam: "sessionPkId", IdParts: 3},
	{Plural: "ldaps", Singular: "ldap", Aliases: []string{"ldap"}, Short: "LDAP servers",
		Columns: []string{"owner", "id", "serverName", "host", "port", "baseDn", "autoSync"}, NameField: "id"},
	{Plural: "syncers", Singular: "syncer", Aliases: []string{"syncer"}, Short: "Syncers",
		Columns: []string{"owner", "name", "type", "host", "database", "table", "isEnabled"}},
	{Plural: "webhooks", Singular: "webhook", Aliases: []string{"webhook"}, Short: "Webhooks",
		Columns: []string{"owner", "name", "url", "method", "events", "isEnabled"}},
	{Plural: "invitations", Singular: "invitation", Aliases: []string{"invitation"}, Short: "Invitation codes",
		Columns: []string{"owner", "name", "code", "quota", "usedCount", "application", "state"}},
	{Plural: "resources", Singular: "resource", Aliases: []string{"resource"}, Short: "Uploaded files",
		Columns: []string{"owner", "name", "user", "provider", "fileType", "fileSize", "url"}},
	{Plural: "products", Singular: "product", Aliases: []string{"product"}, Short: "Products",
		Columns: []string{"owner", "name", "displayName", "price", "currency", "quantity", "state"}},
	{Plural: "orders", Singular: "order", Aliases: []string{"order"}, Short: "Orders",
		Columns: []string{"owner", "name", "user", "products", "price", "currency", "state", "createdTime"}},
	{Plural: "payments", Singular: "payment", Aliases: []string{"payment"}, Short: "Payments",
		Columns: []string{"owner", "name", "user", "provider", "price", "currency", "state", "createdTime"}},
	{Plural: "transactions", Singular: "transaction", Aliases: []string{"transaction"}, Short: "Transactions",
		Columns: []string{"owner", "name", "user", "category", "type", "amount", "currency", "state", "createdTime"}},
	{Plural: "plans", Singular: "plan", Aliases: []string{"plan"}, Short: "Subscription plans",
		Columns: []string{"owner", "name", "displayName", "price", "currency", "period", "role", "isEnabled"}},
	{Plural: "pricings", Singular: "pricing", Aliases: []string{"pricing"}, Short: "Pricings",
		Columns: []string{"owner", "name", "displayName", "application", "plans", "isEnabled"}},
	{Plural: "subscriptions", Singular: "subscription", Aliases: []string{"subscription", "subs"}, Short: "Subscriptions",
		Columns: []string{"owner", "name", "user", "plan", "pricing", "startTime", "endTime", "state"}},
	{Plural: "coupons", Singular: "coupon", Aliases: []string{"coupon"}, Short: "Coupons"},
	{Plural: "keys", Singular: "key", Aliases: []string{"key"}, Short: "API keys"},
	{Plural: "forms", Singular: "form", Aliases: []string{"form"}, Short: "Forms"},
	{Plural: "tickets", Singular: "ticket", Aliases: []string{"ticket"}, Short: "Tickets"},
	{Plural: "agents", Singular: "agent", Aliases: []string{"agent"}, Short: "Agents"},
	{Plural: "servers", Singular: "server", Aliases: []string{"server"}, Short: "MCP servers"},
	{Plural: "entries", Singular: "entry", Aliases: []string{"entry"}, Short: "Entries"},
	{Plural: "sites", Singular: "site", Aliases: []string{"site"}, Short: "Gateway sites"},
	{Plural: "rules", Singular: "rule", Aliases: []string{"rule"}, Short: "Gateway rules"},
	{Plural: "records", Singular: "record", Aliases: []string{"record", "logs"}, Short: "Audit records of the API calls", ListOnly: true,
		Columns: []string{"id", "createdTime", "organization", "user", "method", "requestUri", "action", "statusCode"}},
}

func (r *resource) columns() []string {
	if len(r.Columns) > 0 {
		return r.Columns
	}
	return defaultColumns
}

func (r *resource) nameField() string {
	if r.NameField != "" {
		return r.NameField
	}
	return "name"
}

func (r *resource) idParam() string {
	if r.IdParam != "" {
		return r.IdParam
	}
	return "id"
}

func (r *resource) defaultOwner(organization string) string {
	if r.AdminOwned {
		return "admin"
	}
	return organization
}

// id turns a name into the id of the object, a name that already has the owner is kept.
func (r *resource) id(owner, name string) string {
	parts := r.IdParts
	if parts == 0 {
		parts = 2
	}
	if strings.Count(name, "/") >= parts-1 {
		return name
	}
	return owner + "/" + name
}
