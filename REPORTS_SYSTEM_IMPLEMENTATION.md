# Comprehensive Reports System Implementation

## Overview

A comprehensive reports system has been implemented for the Integrated Outbreak System (IOS) with role-based access control, filtering capabilities, and data visualization. The system supports multiple disease types (VHF, Measles, Polio, Mpox) and provides both general and CIF-specific reporting.

## Key Features

### 1. Role-Based Access Control
- **Full Access**: Admin users can view all data across all facilities and districts
- **Facility Access**: Users restricted to their assigned facility
- **District Access**: Users restricted to their assigned district
- **Limited Access**: Basic access with restrictions

### 2. Filtering Capabilities
- Date range filtering (start/end dates)
- Outbreak-specific filtering
- Facility and district filtering
- Patient type filtering (case, suspect, contact)
- Outcome filtering (recovered, died, in treatment)
- Treatment site filtering
- Symptoms and treatment filtering

### 3. Report Types
- **General Reports**: Overview dashboard, trend analysis, demographics
- **CIF Reports**: Disease-specific Case Investigation Forms
  - VHF CIF Data
  - Measles CIF Data
  - Polio CIF Data
  - Mpox CIF Data
- **Specialized Reports**:
  - Treatment Reports (protocols, outcomes)
  - Laboratory Reports (test results, specimens)
  - Geographic Reports (distribution, hotspots)

### 4. Data Visualization
- Interactive charts using Chart.js
- Multiple chart types (line, bar, pie, doughnut)
- Real-time data updates via AJAX
- Responsive design for mobile devices

## Technical Implementation

### Files Created/Modified

#### 1. Core Reports Logic
- `internal/reports/reports.go` - Main reports functionality
  - Role-based access control
  - Filter parsing and application
  - Report generation functions
  - API endpoints for AJAX data loading

#### 2. Templates
- `ui/html/reports_dashboard.html` - Main reports dashboard
  - Modern, responsive design
  - Filter panel with comprehensive options
  - Quick stats cards
  - Report category cards
  - Interactive charts

- `ui/html/report_view.html` - Individual report view
  - Filter summary display
  - Key metrics cards
  - Multiple chart sections
  - Data tables
  - Disease-specific sections

#### 3. Routes
- `internal/routes/routes.go` - Updated with new report routes
  - Dashboard route (`/reports`)
  - Report generation route (`/reports/generate`)
  - CIF-specific routes (`/reports/cif/:type`)
  - API endpoints for data loading
  - Export functionality

### Database Integration

The system integrates with existing database tables:
- `clients` - Patient data
- `outbreaks` - Outbreak information
- `facilities` - Facility data
- `districts` - District data
- `employee` - Employee assignments
- `users` - User authentication
- Disease-specific tables (VHF, Measles, Polio, Mpox CIF data)

### Security Features

1. **Authentication**: All routes require user authentication
2. **Authorization**: Role-based access control using RBAC system
3. **Data Filtering**: Automatic application of access restrictions
4. **SQL Injection Prevention**: Parameterized queries
5. **Session Management**: Secure session handling

## User Roles and Access

### VHF Lab Technician
- **Special Access**: Can only view CIF data for their assigned district
- **Restrictions**: Limited to district-level data
- **Features**: District-specific CIF reports

### Reports Role
- **Access**: Can view reports based on their facility/district assignment
- **Features**: Full reporting capabilities within their scope

### Admin Role
- **Access**: Full system access
- **Features**: All reports and data across all facilities

### Case Manager
- **Access**: Case-related reports within their scope
- **Features**: Case management and tracking reports

### Data Analyst
- **Access**: Analytical reports and data exports
- **Features**: Advanced analytics and trend analysis

## API Endpoints

### Quick Stats
- `GET /reports/api/stats`
- Returns dashboard statistics (total cases, active cases, recovered, deaths)

### Chart Data
- `GET /reports/api/chart-data?type=<chart_type>`
- Returns chart data for various visualization types
- Supported types: trends, outcomes, facilities, age_groups

### Table Data
- `GET /reports/api/table-data`
- Returns paginated table data with filtering

### Export
- `POST /reports/export`
- Handles report export functionality (CSV, Excel, PDF)

## Usage Instructions

### Accessing Reports
1. Navigate to `/reports` (requires reports permission)
2. Use the filter panel to set criteria
3. Select report type from the available categories
4. View generated reports with charts and data tables

### Filtering Data
1. Set date range for the report period
2. Select specific outbreak if needed
3. Choose facility/district restrictions
4. Filter by patient type and outcome
5. Apply additional filters as needed

### Generating Reports
1. Select report category (General, CIF, Treatment, Lab, Geographic)
2. Choose specific report type
3. Apply filters
4. Click "Generate Report" to view results

### Exporting Data
1. Generate desired report
2. Click "Export" button
3. Choose export format (CSV, Excel, PDF)
4. Download generated file

## Future Enhancements

### Planned Features
1. **Real-time Updates**: WebSocket integration for live data updates
2. **Advanced Analytics**: Machine learning insights and predictions
3. **Custom Dashboards**: User-configurable dashboard layouts
4. **Scheduled Reports**: Automated report generation and distribution
5. **Mobile App**: Native mobile application for field workers

### Technical Improvements
1. **Caching**: Redis integration for improved performance
2. **Data Warehouse**: Dedicated analytics database
3. **API Versioning**: RESTful API with versioning
4. **Microservices**: Service-oriented architecture
5. **Containerization**: Docker deployment support

## Configuration

### Environment Variables
```bash
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ios_db
DB_USER=ios_user
DB_PASSWORD=ios_password

# Application settings
APP_PORT=3000
APP_ENV=development
LOG_LEVEL=info
```

### Database Permissions
Ensure the database user has appropriate permissions:
```sql
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ios_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ios_user;
```

## Troubleshooting

### Common Issues

1. **Access Denied Errors**
   - Check user role assignments
   - Verify RBAC permissions
   - Ensure proper facility/district assignment

2. **Empty Reports**
   - Verify date range settings
   - Check filter criteria
   - Ensure data exists for selected criteria

3. **Performance Issues**
   - Check database indexes
   - Monitor query execution times
   - Consider implementing caching

4. **Chart Loading Issues**
   - Check JavaScript console for errors
   - Verify Chart.js library loading
   - Ensure AJAX endpoints are accessible

### Debug Mode
Enable debug logging by setting:
```bash
LOG_LEVEL=debug
```

## Support

For technical support or feature requests, please contact the development team or create an issue in the project repository.

---

**Version**: 1.0.0  
**Last Updated**: January 2025  
**Compatibility**: Go 1.21+, PostgreSQL 12+  
**Browser Support**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+ 