package reports

import (
	"case/internal/config"
	"case/internal/handlers"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func ReportHome(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	userID, userName := handlers.GetUser(c, sl, store)

	if userID == 0 {
		return c.Redirect("/login", 302)
	}

	// load page
	data := handlers.NewTemplateData(c, store)
	data.User = userName
	data.Form = map[string]string{"Title": "Reports Page", "UserID": userName}
	return handlers.GenerateHTML(c, db, data, "reports")
}

func ReportView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	userID, userName := handlers.GetUser(c, sl, store)

	if userID == 0 {
		return c.Redirect("/login", 302)
	}

	data := handlers.NewTemplateData(c, store)
	data.User = userName
	data.Form = getClinicalSummary(c, db)
	// load page

	return handlers.GenerateHTML(c, db, data, "reportview")
}

func getClinicalSummary(c *fiber.Ctx, db *sql.DB) map[string]string {
	summary := make(map[string]string)

	labelSex := map[string]string{"1": "Male", "2": "Female"}
	table1, _ := GenerateHTMLSummary(c.Context(), db, "clients", "gender", labelSex)

	summary["gender_summary"] = table1

	table2, _ := GenerateHTMLSummary(c.Context(), db, "clients", "status", nil)
	summary["status_summary"] = table2

	labelSex = map[string]string{"1": "Ward", "2": "ICU"}
	table3, _ := GenerateHTMLSummary(c.Context(), db, "clients", "adm_ward", labelSex)
	summary["ward_summary"] = table3

	// Example label map (field name -> label)
	labelCondition := map[string]string{
		"1": "Yes",
		"2": "No",
		"3": "Unknown",
	}

	// Example fields array
	fields := []string{"tb", "asplenia", "hep", "diabetes", "hiv", "liver", "malignancy", "heart", "pulmonary", "kidney", "neurologic"}

	table4, _ := GenerateHTMLFrequencySummary(c.Context(), db, "clients", fields, labelCondition)
	summary["conditions_summary"] = table4

	return summary
}
