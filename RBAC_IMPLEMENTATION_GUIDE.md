# Enhanced User Management & RBAC System Implementation Guide

## Overview

This document outlines a comprehensive Role-Based Access Control (RBAC) system to replace your current basic user management system. The new system provides modern security features, granular permissions, and better user organization.

## Current System Limitations

### What You Have:
- Basic user authentication with sessions
- Simple function-level permissions via `user_right` table
- Limited role management
- No organizational structure
- Basic audit logging

### Issues with Current System:
1. **No Role-Based Access Control**: Only function-level permissions
2. **Weak Security**: Basic password handling, no modern security features
3. **No User Groups/Departments**: Can't organize users by organizational structure
4. **Limited Audit Trail**: Basic user logging
5. **No Permission Inheritance**: Each user needs individual permissions
6. **No Multi-Factor Authentication**: Single factor authentication only

## New RBAC System Features

### 1. Enhanced User Management
- **Modern Password Security**: Salted SHA-256 hashing
- **Account Lockout**: Protection against brute force attacks
- **Password Expiration**: Configurable password policies
- **User Status Management**: Active/inactive/locked states
- **Department Organization**: Users organized by departments
- **Comprehensive User Profiles**: Full name, email, contact info

### 2. Role-Based Access Control
- **Predefined Roles**: Super Admin, Admin, Manager, Data Entry, Viewer, Lab Tech, Surveillance Officer
- **Granular Permissions**: Resource-based permissions (create, read, update, delete, export)
- **Permission Inheritance**: Users inherit permissions through roles
- **Flexible Role Assignment**: Users can have multiple roles
- **Resource-Based Access**: Permissions tied to specific system resources

### 3. Enhanced Security
- **Session Management**: Database-backed session tracking
- **Audit Logging**: Comprehensive audit trail for all actions
- **IP Tracking**: Log user IP addresses and user agents
- **Permission Caching**: Optimized permission checking
- **Access Control Middleware**: Easy-to-use permission checking

### 4. Organizational Structure
- **Departments**: Organize users by organizational units
- **Hierarchical Access**: Department-based access control
- **Scalable Design**: Easy to add new departments and roles

## Implementation Steps

### Step 1: Database Migration
Run the migration script to create the new RBAC tables:

```sql
-- Run the migration file: migrations/012_create_rbac_system.sql
```

This creates:
- `departments` - Organizational departments
- `enhanced_users` - Enhanced user accounts
- `roles` - System roles
- `permissions` - System permissions
- `user_roles` - User-role assignments
- `role_permissions` - Role-permission assignments
- `user_sessions` - Session management
- `audit_logs` - Comprehensive audit trail

### Step 2: Update Application Code

#### 2.1 Update Routes
Replace the current authentication middleware with the new RBAC middleware:

```go
// Old way
appGroup.Use(AuthRequired(store))

// New way
appGroup.Use(middleware.EnhancedAuthRequired())

// For specific permissions
appGroup.Get("/users", middleware.PermissionRequired("users", "read"), userHandler.ListUsers)
appGroup.Post("/users", middleware.PermissionRequired("users", "create"), userHandler.CreateUser)

// For specific roles
appGroup.Get("/admin", middleware.RoleRequired("admin"), adminHandler.Dashboard)
```

#### 2.2 Update Handlers
Use the new permission checking in your handlers:

```go
// Check permissions in handlers
if !middleware.UserHasPermission(c, "vhf_patients", "create") {
    return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
}

// Get current user
userID, _ := middleware.GetCurrentUserID(c)
```

#### 2.3 Add Audit Logging
Log important actions:

```go
auditLog := &models.AuditLog{
    UserID:     sql.NullInt64{Int64: int64(userID), Valid: true},
    Action:     "create",
    Resource:   "vhf_patients",
    ResourceID: sql.NullInt64{Int64: int64(patientID), Valid: true},
    Details:    "Created new VHF patient record",
    IPAddress:  c.IP(),
    UserAgent:  c.Get("User-Agent"),
    CreatedAt:  time.Now(),
}
models.LogAuditEvent(c.Context(), db, auditLog)
```

### Step 3: User Interface Updates

#### 3.1 User Management Interface
Create a modern user management interface with:
- User listing with pagination and filtering
- User creation/editing forms
- Role assignment interface
- Department management
- User status management

#### 3.2 Permission-Based UI
Show/hide UI elements based on permissions:

```javascript
// Check if user has permission to create VHF patients
if (userPermissions.includes('vhf_patients:create')) {
    showCreateButton();
}

// Check if user has admin role
if (userRoles.includes('admin')) {
    showAdminPanel();
}
```

### Step 4: Migration Strategy

#### 4.1 Data Migration
Migrate existing users to the new system:

```sql
-- Migrate existing users (example)
INSERT INTO enhanced_users (username, email, first_name, last_name, password_hash, password_salt, is_active)
SELECT 
    user_name,
    user_name || '@system.local',
    user_name,
    user_name,
    'migrated_password_hash',
    'migrated_salt',
    true
FROM users
WHERE user_name IS NOT NULL;
```

#### 4.2 Permission Mapping
Map existing permissions to new RBAC system:

```sql
-- Map existing user_right to new permissions (example)
-- This requires analysis of your current meta table structure
```

## Predefined Roles and Permissions

### Roles
1. **Super Admin**: Full system access
2. **Admin**: Management access (except user management)
3. **Manager**: Oversight access (read, update, export)
4. **Data Entry**: Create and read permissions
5. **Viewer**: Read-only access
6. **Lab Technician**: Laboratory-specific access
7. **Surveillance Officer**: Surveillance-specific access

### Resources
- `users` - User management
- `vhf_patients` - VHF patient records
- `reports` - System reports
- `outbreaks` - Outbreak management
- `employees` - Employee records
- `facilities` - Health facilities
- `laboratory` - Laboratory results
- `surveillance` - Surveillance data

### Actions
- `create` - Create new records
- `read` - View records
- `update` - Modify records
- `delete` - Delete records
- `export` - Export data
- `import` - Import data

## Security Best Practices

### 1. Password Security
- Use strong password policies
- Implement password expiration
- Enable account lockout after failed attempts
- Use secure password hashing (already implemented)

### 2. Session Management
- Implement session timeout
- Track active sessions
- Allow users to view and terminate their sessions
- Log session events

### 3. Access Control
- Follow principle of least privilege
- Regular permission audits
- Role-based access reviews
- Monitor unauthorized access attempts

### 4. Audit and Monitoring
- Comprehensive audit logging
- Regular audit log reviews
- Monitor suspicious activities
- Automated alerts for security events

## API Endpoints

### User Management
```
GET    /api/users              - List users (with pagination/filtering)
POST   /api/users              - Create user
GET    /api/users/:id          - Get user details
PUT    /api/users/:id          - Update user
DELETE /api/users/:id          - Delete user
```

### Role Management
```
GET    /api/roles              - List roles
POST   /api/roles              - Create role
GET    /api/roles/:id          - Get role details
PUT    /api/roles/:id          - Update role
DELETE /api/roles/:id          - Delete role
```

### Permission Management
```
GET    /api/permissions        - List permissions
GET    /api/users/:id/permissions - Get user permissions
POST   /api/roles/:id/permissions - Assign permissions to role
```

### Audit Logs
```
GET    /api/audit-logs         - List audit logs (with filtering)
GET    /api/audit-logs/:id     - Get audit log details
```

## Implementation Timeline

### Phase 1 (Week 1-2): Database and Core System
- [ ] Run database migration
- [ ] Implement core RBAC models
- [ ] Create enhanced middleware
- [ ] Basic user management handlers

### Phase 2 (Week 3-4): Authentication and Authorization
- [ ] Update authentication system
- [ ] Implement permission checking
- [ ] Add audit logging
- [ ] Update existing routes

### Phase 3 (Week 5-6): User Interface
- [ ] Create user management UI
- [ ] Implement role management interface
- [ ] Add permission-based UI controls
- [ ] Create audit log viewer

### Phase 4 (Week 7-8): Testing and Migration
- [ ] Comprehensive testing
- [ ] Data migration from old system
- [ ] User training
- [ ] Go-live and monitoring

## Benefits of New System

### 1. Security
- **Modern Security**: Salted password hashing, account lockout
- **Granular Control**: Resource-based permissions
- **Audit Trail**: Comprehensive logging of all actions
- **Session Management**: Better session control

### 2. Usability
- **Role-Based**: Easy to manage user access through roles
- **Organizational**: Users organized by departments
- **Scalable**: Easy to add new roles and permissions
- **Flexible**: Support for multiple roles per user

### 3. Management
- **Centralized**: All user management in one place
- **Auditable**: Complete audit trail for compliance
- **Maintainable**: Clean, well-structured code
- **Extensible**: Easy to add new features

### 4. Compliance
- **Audit Logging**: Meets regulatory requirements
- **Access Control**: Proper separation of duties
- **User Management**: Comprehensive user lifecycle management
- **Security**: Modern security practices

## Next Steps

1. **Review the Implementation**: Go through the provided code and documentation
2. **Plan Migration**: Create a detailed migration plan
3. **Test in Development**: Implement and test in development environment
4. **User Training**: Train administrators on new system
5. **Go-Live**: Deploy to production with proper monitoring

## Support and Maintenance

### Regular Tasks
- Review audit logs weekly
- Update user permissions monthly
- Review and update roles quarterly
- Security assessment annually

### Monitoring
- Monitor failed login attempts
- Track permission usage
- Review audit logs for anomalies
- Monitor system performance

This enhanced RBAC system will provide your organization with a modern, secure, and scalable user management solution that meets current security standards and provides the flexibility needed for future growth. 