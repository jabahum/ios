# Outbreak Assignment and Employee Management Implementation Guide

## Overview

This guide covers the implementation of comprehensive outbreak assignment, patient management roles, employee management, and password change functionality for your iOS system.

## Features Implemented

### 1. Outbreak Assignment System
- **User-Outbreak Assignment**: Assign users to specific outbreaks
- **Outbreak Access Control**: Users can only see outbreaks they're assigned to
- **Assignment Management**: Add/remove users from outbreaks
- **Assignment History**: Track who assigned users and when

### 2. Patient Management Roles
- **Role Types**: Registration, Admission, Discharge, Care
- **Granular Permissions**: Outbreak-specific and facility-specific roles
- **Permission Checking**: Real-time permission validation
- **Role Assignment**: Assign multiple roles to users

### 3. Enhanced Employee Management
- **Complete Employee Records**: Full employee information management
- **Department and Title Management**: Structured organizational hierarchy
- **Employee Assignments**: Flexible assignment system for facilities, outbreaks, departments
- **Photo Upload**: Employee photo management
- **Statistics Dashboard**: Employee analytics and reporting
- **Export Functionality**: CSV export of employee data

### 4. Password Management
- **Self-Service Password Change**: Users can change their own passwords
- **Password Strength Validation**: Enforce strong password policies
- **Password Reset**: Email-based password reset functionality
- **Security Audit**: Track password change history

## Database Schema

### New Tables Created

#### 1. `user_outbreaks` - Outbreak Assignment
```sql
CREATE TABLE user_outbreaks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    outbreak_id INTEGER NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES users(user_id),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, outbreak_id)
);
```

#### 2. `patient_management_roles` - Patient Role Assignment
```sql
CREATE TABLE patient_management_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    role_type VARCHAR(50) NOT NULL,
    outbreak_id INTEGER,
    facility_id INTEGER,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id),
    UNIQUE(user_id, role_type, outbreak_id, facility_id)
);
```

#### 3. `password_change_requests` - Password Reset Management
```sql
CREATE TABLE password_change_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    request_token VARCHAR(255) NOT NULL UNIQUE,
    current_password_hash VARCHAR(255) NOT NULL,
    new_password_hash VARCHAR(255) NOT NULL,
    new_password_salt VARCHAR(255) NOT NULL,
    is_approved BOOLEAN DEFAULT false,
    approved_by INTEGER REFERENCES users(user_id),
    approved_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT
);
```

#### 4. Enhanced Employee Tables
```sql
-- Enhanced employee table
ALTER TABLE employee ADD COLUMN employee_code VARCHAR(20) UNIQUE;
ALTER TABLE employee ADD COLUMN employee_title VARCHAR(100);
ALTER TABLE employee ADD COLUMN employee_department VARCHAR(100);
ALTER TABLE employee ADD COLUMN employee_supervisor INTEGER REFERENCES employee(employee_id);
ALTER TABLE employee ADD COLUMN employee_start_date DATE;
ALTER TABLE employee ADD COLUMN employee_end_date DATE;
ALTER TABLE employee ADD COLUMN employee_status VARCHAR(20) DEFAULT 'active';
ALTER TABLE employee ADD COLUMN employee_photo_url VARCHAR(255);
ALTER TABLE employee ADD COLUMN employee_notes TEXT;
ALTER TABLE employee ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE employee ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE employee ADD COLUMN created_by INTEGER REFERENCES users(user_id);
ALTER TABLE employee ADD COLUMN updated_by INTEGER REFERENCES users(user_id);

-- Employee assignments table
CREATE TABLE employee_assignments (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employee(employee_id),
    assignment_type VARCHAR(50) NOT NULL,
    assignment_id INTEGER NOT NULL,
    assignment_name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(user_id)
);

-- Employee departments and titles
CREATE TABLE employee_departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE employee_titles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Implementation Steps

### Step 1: Database Migration
1. Run the migration script:
```bash
psql -h localhost -U postgres -d ios -f migrations/014_add_outbreak_assignment_and_patient_roles.sql
```

2. Verify the tables were created:
```sql
\dt user_outbreaks
\dt patient_management_roles
\dt password_change_requests
\dt employee_assignments
\dt employee_departments
\dt employee_titles
```

### Step 2: Update Go Models
1. The models are already created in:
   - `internal/models/outbreak_assignment.go`
   - `internal/models/employee.go`

2. Update your main application to initialize these services:
```go
// In your main.go or app initialization
userOutbreakService := models.NewUserOutbreakService(db)
patientRoleService := models.NewPatientManagementRoleService(db)
employeeService := models.NewEmployeeService(db)
```

### Step 3: Update Handlers
1. The handlers are created in:
   - `internal/handlers/outbreak_assignment.go`
   - `internal/handlers/employee.go`
   - `internal/handlers/password.go`

2. Initialize handlers in your main application:
```go
outbreakHandler := handlers.NewOutbreakAssignmentHandler(
    userOutbreakService, patientRoleService, userService, outbreakService,
)
employeeHandler := handlers.NewEmployeeHandler(employeeService, userService)
passwordHandler := handlers.NewPasswordHandler(userService)
```

### Step 4: Set Up Routes
1. The routes are defined in:
   - `internal/routes/outbreak_assignment.go`
   - `internal/routes/employee.go`
   - `internal/routes/main.go`

2. Add route setup to your main application:
```go
routes.SetupOutbreakAssignmentRoutes(router, outbreakHandler)
routes.SetupEmployeeRoutes(router, employeeHandler)
routes.SetupAllRoutes(router, handlers)
```

### Step 5: Update UI Templates
1. The UI templates are created:
   - `ui/html/list_employees.html` - Employee listing
   - `ui/html/form_employee.html` - Employee form
   - `ui/html/change_password.html` - Password change form

2. Add navigation links to your main layout:
```html
<a href="/employees" class="nav-link">Employees</a>
<a href="/outbreaks/assignments" class="nav-link">Outbreak Assignments</a>
<a href="/change-password" class="nav-link">Change Password</a>
```

## Usage Examples

### 1. Assigning Users to Outbreaks
```javascript
// Assign users to an outbreak
fetch('/outbreaks/123/assign', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        user_ids: [1, 2, 3]
    })
});
```

### 2. Checking Patient Permissions
```javascript
// Check if user has admission permission for outbreak
fetch('/patient-roles/check-permission?role_type=admission&outbreak_id=123')
    .then(response => response.json())
    .then(data => {
        if (data.has_permission) {
            // Show admission form
        }
    });
```

### 3. Employee Management
```javascript
// Create new employee
const employeeData = {
    employee_fname: 'John',
    employee_lname: 'Doe',
    employee_email: 'john.doe@example.com',
    employee_department: 'Clinical Services',
    employee_title: 'Medical Officer'
};

fetch('/employees/save', {
    method: 'POST',
    body: new FormData(employeeForm)
});
```

### 4. Password Change
```javascript
// Change password
const passwordData = {
    current_password: 'oldPassword123',
    new_password: 'newSecurePassword456!',
    confirm_password: 'newSecurePassword456!',
    csrf_token: 'token-from-cookie'
};

fetch('/change-password', {
    method: 'POST',
    body: new FormData(passwordForm)
});
```

## Security Features

### 1. Authentication & Authorization
- All routes require authentication
- Role-based access control
- Outbreak-specific permissions
- CSRF protection for forms

### 2. Password Security
- Strong password requirements
- Salted password hashing
- Password change audit trail
- Secure password reset tokens

### 3. Data Protection
- Input validation and sanitization
- SQL injection prevention
- XSS protection
- Secure session management

## Default Data

### Employee Departments
- Clinical Services
- Laboratory Services
- Surveillance
- Administration
- Support Services
- Research

### Employee Titles
- Medical Officer
- Nurse
- Laboratory Technician
- Surveillance Officer
- Data Entry Clerk
- Administrator
- Support Staff
- Research Officer

### Patient Management Roles
- `registration` - Patient registration
- `admission` - Patient admission
- `discharge` - Patient discharge
- `care` - Patient care management

## API Endpoints

### Outbreak Assignment
- `GET /outbreaks/:id/assign` - Show assignment form
- `POST /outbreaks/:id/assign` - Assign users to outbreak
- `DELETE /outbreaks/:outbreak_id/users/:user_id` - Remove user from outbreak
- `GET /outbreaks/assignments` - List all assignments
- `GET /outbreaks/my-outbreaks` - Get user's outbreaks

### Patient Roles
- `GET /patient-roles/assign` - Show role assignment form
- `POST /patient-roles/assign` - Assign patient role
- `DELETE /patient-roles/remove` - Remove patient role
- `GET /patient-roles/user/:user_id` - Get user's roles
- `GET /patient-roles/check-permission` - Check permission

### Employee Management
- `GET /employees` - List employees
- `GET /employees/new/:id` - Show employee form
- `POST /employees/save` - Save employee
- `DELETE /employees/:id` - Delete employee
- `GET /employees/:id/details` - Get employee details
- `GET /employees/export` - Export employees
- `GET /employees/statistics` - Get statistics

### Password Management
- `GET /change-password` - Show password change form
- `POST /change-password` - Change password
- `GET /auth/forgot-password` - Show forgot password form
- `POST /auth/forgot-password` - Request password reset
- `GET /auth/reset-password/:token` - Show reset form
- `POST /auth/reset-password/:token` - Reset password

## Testing

### 1. Unit Tests
Create tests for:
- Model functions
- Handler functions
- Service functions

### 2. Integration Tests
Test:
- Database operations
- API endpoints
- Form submissions
- Permission checks

### 3. User Acceptance Testing
Test:
- User workflows
- Permission scenarios
- Error handling
- UI responsiveness

## Monitoring and Maintenance

### 1. Logging
- Log all password changes
- Log outbreak assignments
- Log employee modifications
- Log permission checks

### 2. Performance Monitoring
- Monitor database query performance
- Track API response times
- Monitor user activity

### 3. Regular Maintenance
- Clean up expired password reset tokens
- Archive old employee records
- Update employee statistics
- Backup critical data

## Troubleshooting

### Common Issues

1. **Permission Denied Errors**
   - Check user roles and permissions
   - Verify outbreak assignments
   - Check facility assignments

2. **Password Reset Issues**
   - Verify email configuration
   - Check token expiration
   - Validate email addresses

3. **Employee Code Generation**
   - Check for duplicate names
   - Verify database triggers
   - Check sequence values

### Debug Mode
Enable debug logging:
```go
log.SetLevel(log.DebugLevel)
```

## Next Steps

1. **Email Integration**: Implement email service for password resets
2. **Advanced Reporting**: Create detailed analytics dashboards
3. **Mobile Support**: Develop mobile-responsive interfaces
4. **API Documentation**: Create comprehensive API documentation
5. **Performance Optimization**: Implement caching and optimization
6. **Audit Trail**: Enhanced logging and audit features
7. **Integration**: Connect with external HR systems
8. **Notifications**: Real-time notification system

## Support

For implementation support:
1. Check the database migration logs
2. Verify all dependencies are installed
3. Test with sample data first
4. Review error logs for specific issues
5. Ensure proper database permissions

This implementation provides a robust foundation for outbreak assignment, employee management, and enhanced security features in your iOS system. 