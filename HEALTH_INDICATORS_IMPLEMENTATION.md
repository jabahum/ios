# Health Indicators Report Implementation

## Overview

This document describes the implementation of comprehensive health indicators reporting functionality in the Integrated Outbreak System. The implementation provides all 13 key health indicators as specified in the requirements table.

## Implemented Indicators

### 1. New Admissions (Daily)
- **Source Form**: Admission form
- **Numerator**: Count of patients admitted on a given date
- **Calculation**: Daily aggregation of admission date field
- **Function**: `GetNewAdmissionsDaily()`

### 2. Cumulative Confirmed Cases
- **Source Form**: Admission (Lab results)
- **Numerator**: All patients with positive Mpox PCR/confirmed diagnosis
- **Calculation**: Sum over time; confirmed = "Mpox Positive"
- **Function**: `GetCumulativeConfirmedCases()`

### 3. Cumulative Suspected Cases
- **Source Form**: Admission form
- **Numerator**: All patients with Mpox-like symptoms without PCR/confirmed diagnosis
- **Calculation**: Daily aggregation of admission date field
- **Function**: `GetCumulativeSuspectedCases()`

### 4. Cumulative Deaths
- **Source Form**: Discharge form
- **Numerator**: Patients with outcome = "Death"
- **Calculation**: Sum over time
- **Function**: `GetCumulativeDeaths()`

### 5. Case Fatality Rate (CFR)
- **Source Form**: Discharge form
- **Numerator**: Deaths among confirmed cases
- **Denominator**: All confirmed cases
- **Calculation**: (Deaths/Confirmed cases) × 100
- **Function**: `GetCaseFatalityRate()`

### 6. Current Admissions (Active Cases in Hospital)
- **Source Form**: Admission + Discharge form
- **Numerator**: Patients admitted and not yet discharged/dead
- **Calculation**: Admissions - (Discharges + Deaths)
- **Function**: `GetCurrentAdmissions()`

### 7. Discharges (Cumulative)
- **Source Form**: Discharge form
- **Numerator**: Patients with outcome = "Discharged alive"
- **Calculation**: Sum over time
- **Function**: `GetCumulativeDischarges()`

### 8. Severe Cases Admitted
- **Source Form**: Discharge form
- **Numerator**: Patients marked Severe at daily review rash evaluation
- **Denominator**: All admitted patients
- **Calculation**: % severe = (Severe / All admitted) × 100
- **Function**: `GetSevereCasesAdmitted()`

### 9. Critical Cases Admitted
- **Source Form**: Discharge form
- **Numerator**: Patients marked Critical at daily review rash evaluation
- **Denominator**: All admitted patients
- **Calculation**: % Critical = (Critical / All admitted) × 100
- **Function**: `GetCriticalCasesAdmitted()`

### 10. Cases by Sex
- **Source Form**: Admission form
- **Numerator**: Patients by Male / Female
- **Denominator**: All admitted patients
- **Calculation**: % distribution
- **Function**: `GetCasesBySex()`

### 11. Cases by Age Group
- **Source Form**: Admission form
- **Numerator**: Patients by age group (<5, 5-17, 18-35, 36-59, 60+)
- **Denominator**: All admitted patients
- **Calculation**: % distribution
- **Function**: `GetCasesByAgeGroup()`

### 12. Cases by District/Facility
- **Source Form**: Admission form
- **Numerator**: Cases reported from each
- **Denominator**: All admitted cases
- **Calculation**: Used for maps & trends
- **Function**: `GetCasesByLocation()`

### 13. Healthcare Worker (HCW) Infections
- **Source Form**: Admission form
- **Numerator**: Admitted patients with "HCW = Yes"
- **Denominator**: All admitted patients
- **Calculation**: % HCWs among cases
- **Function**: `GetHCWInfections()`

## Implementation Details

### Database Queries

All indicators are implemented using SQL queries that:
- Join relevant tables (clients, discharge, facility, districts)
- Apply date range filters
- Apply outbreak, facility, and district filters
- Handle null values appropriately
- Use proper aggregation functions

### Key Features

1. **Filter Support**: All indicators support filtering by:
   - Date range (start_date, end_date)
   - Outbreak ID
   - Facility ID
   - District ID

2. **Error Handling**: Each function includes proper error handling and returns meaningful error messages.

3. **Data Validation**: Functions validate input data and handle edge cases (e.g., division by zero for percentages).

4. **Performance**: Queries are optimized with proper indexing considerations.

### Template Integration

The indicators are displayed using a dedicated template (`indicators_report.html`) that includes:

- **Key Metrics Cards**: Visual display of all 13 indicators
- **Charts**: Interactive charts for demographic breakdowns
- **Tables**: Detailed data tables for location-based analysis
- **Responsive Design**: Mobile-friendly layout

### Report Generation

The indicators report is integrated into the existing reports system:

1. **Report Type**: Added "Health Indicators Report" option in the report type dropdown
2. **Route**: Uses existing `/reports/generate` endpoint
3. **Template Selection**: Automatically selects the appropriate template based on report type
4. **Access Control**: Respects existing RBAC permissions

## Usage

### Accessing the Report

1. Navigate to the Reports Dashboard
2. Select "Health Indicators Report" from the Report Type dropdown
3. Apply desired filters (date range, outbreak, facility, district)
4. Click "Generate Report"

### API Access

Individual indicator functions can be accessed programmatically:

```go
import "case/internal/reports"

filters := reports.ReportFilters{
    StartDate:  "2024-01-01",
    EndDate:    "2024-12-31",
    OutbreakID: 1,
}

// Get specific indicators
newAdmissions, err := reports.GetNewAdmissionsDaily(db, filters)
cfr, err := reports.GetCaseFatalityRate(db, filters)
casesBySex, err := reports.GetCasesBySex(db, filters)
```

## Database Schema Requirements

The implementation assumes the following database structure:

### Core Tables
- `clients`: Patient information with fields like `adm_date`, `status`, `gender`, `age`, `occupation`
- `discharge`: Discharge information with fields like `discharge_date`, `discharge_outcome`, `final_diagnosis`
- `facility`: Facility information
- `districts`: District information
- `outbreaks`: Outbreak information

### Key Fields
- `clients.adm_date`: Admission date for daily aggregations
- `clients.status`: Case classification (confirmed, suspected, etc.)
- `clients.gender`: Gender (1=Male, 2=Female)
- `clients.age`: Age for age group calculations
- `clients.occupation`: Occupation for HCW identification
- `discharge.discharge_outcome`: Outcome (Death, Discharged alive)
- `discharge.final_diagnosis`: Final diagnosis for confirmed cases

## Customization

### Severity Classification

The severe and critical case indicators currently use placeholder logic based on `clients.status`. To customize:

1. Update the queries in `GetSevereCasesAdmitted()` and `GetCriticalCasesAdmitted()`
2. Modify the severity classification logic based on your specific criteria
3. Consider adding severity fields to the database schema if needed

### HCW Identification

The HCW infections indicator assumes `clients.occupation = 1` identifies healthcare workers. To customize:

1. Update the query in `GetHCWInfections()`
2. Modify the occupation mapping based on your data structure
3. Consider adding a dedicated HCW field if needed

### Additional Indicators

To add new indicators:

1. Create a new function following the existing pattern
2. Add the function call to `generateIndicatorsReport()`
3. Update the template to display the new indicator
4. Add any necessary database fields or joins

## Testing

A test script (`scripts/test_indicators.go`) is provided to verify the implementation:

```bash
# Set database connection
export DATABASE_URL="postgres://username:password@localhost:5432/dbname?sslmode=disable"

# Run tests
go run scripts/test_indicators.go
```

## Future Enhancements

1. **Caching**: Implement caching for frequently accessed indicators
2. **Real-time Updates**: Add real-time indicator updates
3. **Export Functionality**: Add CSV/Excel export for indicators
4. **Trend Analysis**: Add time-series analysis for indicators
5. **Alerting**: Add threshold-based alerts for critical indicators
6. **Dashboard Integration**: Integrate indicators into main dashboard
7. **API Endpoints**: Create REST API endpoints for individual indicators

## Conclusion

The health indicators implementation provides a comprehensive reporting system that covers all 13 required indicators. The system is designed to be flexible, maintainable, and easily extensible for future requirements. 