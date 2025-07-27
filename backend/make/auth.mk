# =============================================================================
# AUTHENTICATION MODULE - make/auth.mk
# =============================================================================
# Authentication, session management, and user operations

.PHONY: auth-setup auth-create-admin auth-create-user auth-list-users auth-test

# =============================================================================
# AUTHENTICATION SETUP
# =============================================================================

auth-setup: ## 🛠️  Setup authentication system
	@echo "$(GREEN)🔐 Setting up authentication system...$(RESET)"
	@echo "$(BLUE)Session storage: In-memory (Phase 1)$(RESET)"
	@echo "$(YELLOW)💡 SQLite sessions available in Phase 2$(RESET)"
	@echo "$(GREEN)✅ Authentication system ready$(RESET)"

auth-verify: ## 🛠️  Verify authentication configuration
	@echo "$(YELLOW)🔍 Verifying authentication configuration...$(RESET)"
	@go run cmd/tools/auth/verify.go
	@echo "$(GREEN)✅ Authentication configuration verified$(RESET)"

# =============================================================================
# USER MANAGEMENT
# =============================================================================

auth-create-admin: ## 🏢 Create admin user
	@echo "$(GREEN)👤 Creating admin user...$(RESET)"
	@read -p "Email: " email && \
	read -s -p "Password: " password && echo && \
	read -p "Company: " company && \
	read -p "Tenant ID: " tenant && \
	go run cmd/tools/user/create.go \
		--email="$$email" \
		--password="$$password" \
		--role=admin \
		--company="$$company" \
		--tenant="$$tenant" && \
	echo "$(GREEN)✅ Admin user created$(RESET)"

auth-create-user: ## 🏢 Create regular user
	@echo "$(GREEN)👤 Creating user...$(RESET)"
	@read -p "Email: " email && \
	read -s -p "Password: " password && echo && \
	read -p "Company: " company && \
	read -p "Tenant ID: " tenant && \
	read -p "Role (user/operator/manager): " role && \
	go run cmd/tools/user/create.go \
		--email="$$email" \
		--password="$$password" \
		--role="$$role" \
		--company="$$company" \
		--tenant="$$tenant" && \
	echo "$(GREEN)✅ User created$(RESET)"

auth-list-users: ## 🏢 List all users
	@echo "$(BLUE)👥 System Users$(RESET)"
	@echo "==============="
	@go run cmd/tools/users/list.go

auth-delete-user: ## 🏢 Delete user
	@echo "$(RED)⚠️  Deleting user...$(RESET)"
	@read -p "User email to delete: " email && \
	read -p "Are you sure? (y/N): " confirm && \
	[ "$$confirm" = "y" ] && \
	go run cmd/tools/user/delete.go --email="$$email" && \
	echo "$(GREEN)✅ User deleted$(RESET)"

# =============================================================================
# SESSION MANAGEMENT
# =============================================================================

auth-sessions: ## 📊 List active sessions
	@echo "$(BLUE)📊 Active Sessions$(RESET)"
	@echo "=================="
	@go run cmd/tools/sessions/list.go

auth-sessions-cleanup: ## 🧹 Clean expired sessions
	@echo "$(YELLOW)🧹 Cleaning expired sessions...$(RESET)"
	@go run cmd/tools/sessions/cleanup.go
	@echo "$(GREEN)✅ Expired sessions cleaned$(RESET)"

auth-revoke-session: ## 🏢 Revoke specific session
	@echo "$(RED)⚠️  Revoking session...$(RESET)"
	@read -p "Session ID to revoke: " session_id && \
	go run cmd/tools/session/revoke.go --session="$$session_id" && \
	echo "$(GREEN)✅ Session revoked$(RESET)"

auth-revoke-user-sessions: ## 🏢 Revoke all sessions for user
	@echo "$(RED)⚠️  Revoking all sessions for user...$(RESET)"
	@read -p "User email: " email && \
	go run cmd/tools/sessions/revoke-user.go --email="$$email" && \
	echo "$(GREEN)✅ All user sessions revoked$(RESET)"

# =============================================================================
# SECURITY OPERATIONS
# =============================================================================

auth-change-password: ## 🏢 Change user password
	@echo "$(YELLOW)🔑 Changing user password...$(RESET)"
	@read -p "User email: " email && \
	read -s -p "New password: " password && echo && \
	go run cmd/tools/password/change.go \
		--email="$$email" \
		--password="$$password" && \
	echo "$(GREEN)✅ Password changed$(RESET)"

auth-reset-password: ## 🏢 Reset user password (generates temporary)
	@echo "$(YELLOW)🔄 Resetting user password...$(RESET)"
	@read -p "User email: " email && \
	go run cmd/tools/password/reset.go --email="$$email" && \
	echo "$(GREEN)✅ Password reset, temporary password generated$(RESET)"

auth-lock-user: ## 🏢 Lock user account
	@echo "$(RED)🔒 Locking user account...$(RESET)"
	@read -p "User email: " email && \
	read -p "Reason: " reason && \
	go run cmd/tools/user/lock.go \
		--email="$$email" \
		--reason="$$reason" && \
	echo "$(GREEN)✅ User account locked$(RESET)"

auth-unlock-user: ## 🏢 Unlock user account
	@echo "$(GREEN)🔓 Unlocking user account...$(RESET)"
	@read -p "User email: " email && \
	go run cmd/tools/user/unlock.go --email="$$email" && \
	echo "$(GREEN)✅ User account unlocked$(RESET)"

# =============================================================================
# TENANT PERMISSIONS
# =============================================================================

auth-tenant-permissions: ## 🏢 Show user tenant permissions
	@echo "$(BLUE)🏢 Tenant Permissions$(RESET)"
	@echo "====================="
	@read -p "User email: " email && \
	go run cmd/tools/tenant/show-permissions.go --email="$$email"

auth-grant-tenant-access: ## 🏢 Grant user access to tenant
	@echo "$(GREEN)✅ Granting tenant access...$(RESET)"
	@read -p "User email: " email && \
	read -p "Tenant ID: " tenant && \
	read -p "Role (user/operator/manager/admin): " role && \
	go run cmd/tools/tenant/grant-access.go \
		--email="$$email" \
		--tenant="$$tenant" \
		--role="$$role" && \
	echo "$(GREEN)✅ Tenant access granted$(RESET)"

auth-revoke-tenant-access: ## 🏢 Revoke user access to tenant
	@echo "$(RED)❌ Revoking tenant access...$(RESET)"
	@read -p "User email: " email && \
	read -p "Tenant ID: " tenant && \
	go run cmd/tools/tenant/revoke-access.go \
		--email="$$email" \
		--tenant="$$tenant" && \
	echo "$(GREEN)✅ Tenant access revoked$(RESET)"

# =============================================================================
# TESTING
# =============================================================================

auth-test: ## 🧪 Run authentication tests
	@echo "$(YELLOW)🧪 Running authentication tests...$(RESET)"
	@go test -v ./internal/auth/...
	@echo "$(GREEN)✅ Authentication tests complete$(RESET)"

auth-test-login: ## 🧪 Test login functionality
	@echo "$(YELLOW)🧪 Testing login functionality...$(RESET)"
	@read -p "Test email: " email && \
	read -s -p "Test password: " password && echo && \
	curl -X POST http://localhost:$(API_PORT)/auth/login \
		-H "Content-Type: application/json" \
		-d "{\"email\":\"$$email\",\"password\":\"$$password\"}" && \
	echo "" && echo "$(GREEN)✅ Login test complete$(RESET)"

auth-test-endpoints: ## 🧪 Test authentication endpoints
	@echo "$(YELLOW)🧪 Testing authentication endpoints...$(RESET)"
	@echo "Testing health endpoint..."
	@curl -s http://localhost:$(API_PORT)/health | jq .
	@echo ""
	@echo "Testing protected endpoint (should fail without auth)..."
	@curl -s http://localhost:$(API_PORT)/api/v1/customers || echo "✅ Correctly rejected"

# =============================================================================
# SESSION STORAGE UPGRADES
# =============================================================================

auth-upgrade-sqlite: ## ⚡ Upgrade to SQLite session storage
	@echo "$(YELLOW)⚡ Upgrading to SQLite session storage...$(RESET)"
	@go run cmd/tools/sessions/upgrade-sqlite.go
	@echo "$(GREEN)✅ Upgraded to SQLite session storage$(RESET)"
	@echo "$(BLUE)💡 Restart API server to use SQLite sessions$(RESET)"

auth-upgrade-nats: ## ⚡ Upgrade to NATS session storage
	@echo "$(YELLOW)⚡ Upgrading to NATS session storage...$(RESET)"
	@go run cmd/tools/sessions/upgrade-sessions-nats.go
	@echo "$(GREEN)✅ Upgraded to NATS session storage$(RESET)"
	@echo "$(BLUE)💡 Restart API server to use NATS sessions$(RESET)"

# =============================================================================
# SECURITY MONITORING
# =============================================================================

auth-audit-logins: ## 📊 Show recent login attempts
	@echo "$(BLUE)📊 Recent Login Attempts$(RESET)"
	@echo "========================="
	@go run cmd/tools/auth/audit-logins.go --days=7

auth-audit-failed: ## 📊 Show failed login attempts
	@echo "$(BLUE)📊 Failed Login Attempts$(RESET)"
	@echo "========================="
	@go run cmd/tools/auth/audit-failed-logins.go --days=1

auth-security-report: ## 📊 Generate security report
	@echo "$(BLUE)📊 Security Report$(RESET)"
	@echo "=================="
	@go run cmd/tools/auth/security-report.go

# =============================================================================
# CLEANUP
# =============================================================================

auth-clean: ## 🛠️  Clean authentication artifacts
	@echo "$(YELLOW)🧹 Cleaning authentication artifacts...$(RESET)"
	@rm -rf logs/auth-*.log
	@rm -rf tmp/sessions-*.db
	@echo "$(GREEN)✅ Authentication cleanup complete$(RESET)"

# =============================================================================
# HELP
# =============================================================================

help-auth: ## 📖 Show authentication commands help
	@echo "$(BLUE)Authentication Module Commands$(RESET)"
	@echo "==============================="
	@echo ""
	@echo "$(GREEN)🛠️  SETUP:$(RESET)"
	@echo "  auth-setup            - Setup authentication system"
	@echo "  auth-verify           - Verify authentication config"
	@echo ""
	@echo "$(YELLOW)👤 USER MANAGEMENT:$(RESET)"
	@echo "  auth-create-admin     - Create admin user"
	@echo "  auth-create-user      - Create regular user"
	@echo "  auth-list-users       - List all users"
	@echo "  auth-delete-user      - Delete user"
	@echo ""
	@echo "$(BLUE)📊 SESSIONS:$(RESET)"
	@echo "  auth-sessions         - List active sessions"
	@echo "  auth-sessions-cleanup - Clean expired sessions"
	@echo "  auth-revoke-session   - Revoke specific session"
	@echo ""
	@echo "$(RED)🔐 SECURITY:$(RESET)"
	@echo "  auth-change-password  - Change user password"
	@echo "  auth-reset-password   - Reset user password"
	@echo "  auth-lock-user        - Lock user account"
	@echo "  auth-unlock-user      - Unlock user account"
	@echo ""
	@echo "$(GREEN)🏢 TENANT PERMISSIONS:$(RESET)"
	@echo "  auth-tenant-permissions    - Show user tenant permissions"
	@echo "  auth-grant-tenant-access   - Grant tenant access"
	@echo "  auth-revoke-tenant-access  - Revoke tenant access"
	@echo ""
	@echo "$(BLUE)🧪 TESTING:$(RESET)"
	@echo "  auth-test             - Run authentication tests"
	@echo "  auth-test-login       - Test login functionality"
	@echo "  auth-test-endpoints   - Test auth endpoints"
	@echo ""
	@echo "$(YELLOW)⚡ UPGRADES:$(RESET)"
	@echo "  auth-upgrade-sqlite   - Upgrade to SQLite sessions"
	@echo "  auth-upgrade-nats     - Upgrade to NATS sessions"
