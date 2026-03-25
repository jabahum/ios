package handlers

import (
	"bytes"
	"case/internal/config"
	"case/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

func Discharge(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	//=================

	userID := GetCurrentUser(c, store)

	// Check if user is logged in
	if userID == 0 {
		fmt.Println("unauthorized")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	//=============================

	var formData map[string]interface{}

	if err := c.BodyParser(&formData); err != nil {
		fmt.Println("JSON parsing failed:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var s models.Discharge

	s.DischargeID = int(ParseNullInt2(formData["discharge_id"]).Int64)
	s.ClientID = ParseNullInt2(formData["client_id"])
	s.DischargeDate = ParseNullString2(formData["discharge_date"])
	s.FinalDiagnosis = ParseNullString2(formData["final_diagnosis"])
	s.FinalDiagnosisOther = ParseNullString2(formData["final_diagnosis_other"])
	s.DischargeOutcome = ParseNullString2(formData["discharge_outcome"])
	s.DischargeSeqHeari = ParseNullInt2(formData["discharge_seq_heari"])
	s.DischargeSeqPregn = ParseNullInt2(formData["discharge_seq_pregn"])
	s.DischargeSeqOcula = ParseNullInt2(formData["discharge_seq_ocula"])
	s.DischargeSeqExtre = ParseNullInt2(formData["discharge_seq_extre"])
	s.DischargeSeqArthr = ParseNullInt2(formData["discharge_seq_arthr"])
	s.DischargeSeqNeuro = ParseNullInt2(formData["discharge_seq_neuro"])
	s.DischargeSeqOthers = ParseNullInt2(formData["discharge_seq_others"])
	s.CounsellingProvided = ParseNullString2(formData["counselling_provided"])
	s.DischargingOfficer = ParseNullString2(formData["discharging_officer"])
	s.DischargeFacility = ParseNullString2(formData["discharge_facility"])
	s.DischargeSeqOthersAza = ParseNullString2(formData["discharge_seq_others_aza"])

	s.UpdatedBy.Valid = true
	s.UpdatedBy.Int64 = int64(userID)

	s.UpdatedOn.Valid = true
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02")
	s.UpdatedOn.String = formattedTime

	fmt.Println(s)
	// Check if it's a new record (StatusID == 0)
	if s.DischargeID > 0 {
		s.EnteredBy.Valid = true
		s.EnteredBy.Int64 = int64(userID)

		s.EnteredOn.Valid = true
		s.EnteredOn.String = formattedTime

		s.SetAsExists()
		err := s.Update(c.Context(), db)
		if err != nil {
			fmt.Println("update fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	} else {

		err := s.Insert(c.Context(), db)
		if err != nil {
			fmt.Println("insert fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
	})
}

func GetDischarge(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	userID := GetCurrentUser(c, store)

	// Check if user is logged in
	if userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = "0"
	}

	c_id, err := strconv.Atoi(clientID)
	if err != nil {
		c_id = 0
	}

	discharge, er := models.DischargeByClientID(c.Context(), db, c_id)
	if er != nil {
		fmt.Println(er.Error())
		return nil
	}

	return c.JSON(discharge)
}

func Certificate(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {

	clientID := c.Query("who")
	if clientID == "" {
		clientID = "0"
	}

	c_id, er := strconv.Atoi(clientID)
	if er != nil {
		c_id = 0
	}

	var discharge models.Discharge

	discharge = models.Discharge{}
	discharge.DischargeDate.Valid = true
	discharge.DischargeDate.String = ""

	disc, erx := models.DischargeByClientID(c.Context(), db, c_id)
	if erx != nil {
		fmt.Println("discharge: " + erx.Error())
	} else {
		discharge = *disc
	}

	client, erx := models.ClientByID(c.Context(), db, c_id)
	if erx != nil {
		fmt.Println("client: " + erx.Error())
		//return nil
	}

	facility, erx := models.FacilityByFacilityID(c.Context(), db, int(client.Site.Int64))
	if erx != nil {
		fmt.Println("facility: " + erx.Error())
		//return nil
	}

	// Create a new PDF
	fmt.Println("Starting certificate...")
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	// Get page width and height (A4 landscape: 297mm x 210mm)
	pageWidth, pageHeight := pdf.GetPageSize()

	// Save default line settings
	defaultLineWidth := pdf.GetLineWidth()
	defaultDrawColorR, defaultDrawColorG, defaultDrawColorB := 0, 0, 0 // Default is black

	// Define starting margin
	borderMargin := 10.0

	// 🖤 **Black Border (Outer)**
	pdf.SetLineWidth(3)
	pdf.SetDrawColor(0, 0, 0) // Black
	pdf.Rect(borderMargin, borderMargin, pageWidth-2*borderMargin, pageHeight-2*borderMargin, "D")

	// 💛 **Yellow Border (Middle, touching the black border)**
	pdf.SetLineWidth(2.5)
	pdf.SetDrawColor(255, 204, 0) // Yellow
	pdf.Rect(borderMargin+1.5, borderMargin+1.5, pageWidth-2*(borderMargin+1.5), pageHeight-2*(borderMargin+1.5), "D")

	// ❤️ **Red Border (Inner, touching the yellow border)**
	pdf.SetLineWidth(2)
	pdf.SetDrawColor(255, 0, 0) // Red
	pdf.Rect(borderMargin+3, borderMargin+3, pageWidth-2*(borderMargin+3), pageHeight-2*(borderMargin+3), "D")

	// 🔄 **Reset to Default Line Settings**
	pdf.SetLineWidth(defaultLineWidth)                                        // Reset line width
	pdf.SetDrawColor(defaultDrawColorR, defaultDrawColorG, defaultDrawColorB) // Reset line color to black

	// 📝 **Example Certificate Text (Centered)**
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 0, 0)

	// Load Fonts
	pdf.SetFont("Arial", "B", 10)

	// Add Ministry of Health Logo
	logoPath := "../../ui/static/img/logo.png"
	pdf.Image(logoPath, 20, 15, 30, 0, false, "", 0, "") // Centered Logo
	//pdf.Ln(35)

	// Generate QR Code
	qrLink := "response.health.go.ug/verify/discharges/" + strconv.Itoa(discharge.DischargeID) // Replace with the actual verification link
	qrFile := "qrcode.png"
	err := qrcode.WriteFile(qrLink, qrcode.Medium, 256, qrFile)
	if err != nil {
		fmt.Println(err.Error())
		//panic(err)
	}

	// Add QR Code to PDF
	pdf.Image(qrFile, 240, 15, 30, 30, false, "", 0, "")
	_ = os.Remove(qrFile) // Cleanup QR Code file

	// Ministry of Health Title

	logo_text := "Ministry of Health, Republic of Uganda"

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(260, 20, logo_text, "0", 1, "C", false, 0, "")
	//pdf.Ln(5)

	// Certificate Title
	title_text := "EVD Discharge Certificate"

	pdf.SetFont("Arial", "B", 24)
	pdf.CellFormat(260, 10, title_text, "0", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Certification Text

	body1_text := "This is to certify that"

	pdf.SetFont("Arial", "i", 14)
	pdf.CellFormat(260, 10, body1_text, "0", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Name (Handwriting Style)
	pdf.SetFont("Courier", "B", 22) // Courier gives a handwritten feel
	pdf.CellFormat(260, 12, client.Firstname.String+" "+client.Lastname.String+" ("+client.EtuNo.String+")", "0", 1, "C", false, 0, "")
	pdf.Ln(1)

	// Draw Line Below Name
	pdf.Line(55, pdf.GetY(), 260, pdf.GetY()) // Draw a horizontal line
	pdf.Ln(5)

	// Rest of the Certification Text
	pdf.SetFont("Arial", "", 12)
	text1 := "At the date (" + discharge.DischargeDate.String + ") of issue of this certificate, does not present a risk of infecting other persons after testing negative for Ebola\nVirus Disease. The current state of health does not constitute a danger to the community and can therefore, return to his/her\nhousehold and professional environment to continue his/her normal daily activities"
	pdf.MultiCell(0, 6, text1, "", "C", false)
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 12)
	text2 := `The family, the community, and the authorities are requested to accept his/her in order to promote his/her social re-integration`
	pdf.MultiCell(0, 10, text2, "", "C", false)
	pdf.Ln(5)

	// Draw a box around the last three words

	pdf.SetFont("Arial", "", 14)
	firstwords := `Completed at facility: `

	pdf.SetFont("Arial", "B", 12) // Set bold font
	lastWords := facility.FacilityName.String

	all_words := firstwords + lastWords

	// Draw the box
	pdf.Rect(20, pdf.GetY(), 250, 10, "D") // Adjust the size of the box (190 width, 30 height)

	// Write the text in the box
	pdf.MultiCell(0, 10, all_words, "", "C", false)

	pdf.Ln(10) // Line break after the text

	// Signatures Section

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(140, 10, "___________________________", "0", 0, "C", false, 0, "")
	pdf.CellFormat(140, 10, "___________________________", "0", 1, "C", false, 0, "")

	pdf.CellFormat(140, 10, "Director General of Health Services", "0", 0, "C", false, 0, "")
	pdf.CellFormat(140, 10, "Ebola Treatment Unit Manager", "0", 1, "C", false, 0, "")

	pdf.CellFormat(140, 10, "Ministry of Health", "0", 0, "C", false, 0, "")
	pdf.CellFormat(140, 10, facility.FacilityName.String, "0", 1, "C", false, 0, "")

	pdf.Ln(25)

	// Save PDF to Buffer (In-Memory)
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		fmt.Println("Cert error: ", err.Error())
		//return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed, contact admin"})
	}

	// Set HTTP Headers and Return PDF
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=certificate.pdf")
	c.Set("Content-Length", fmt.Sprintf("%d", buf.Len())) // Set content length for the browser
	return c.Send(buf.Bytes())

}
