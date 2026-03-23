"""Remove duplicate protected /api block from routes.go (now in app_json_api.go)."""
from pathlib import Path

p = Path(__file__).resolve().parents[1] / "internal/routes/routes.go"
s = p.read_text(encoding="utf-8")
start = s.find("\t// Lab Sample Types API routes (protected - require authentication)")
if start < 0:
    raise SystemExit("start not found")
end = s.find("\tprotected.Delete(\"/api/alerts/:id\", func(c *fiber.Ctx) error {", start)
if end < 0:
    raise SystemExit("alerts delete not found")
end = s.find("\n\t})\n", end)
if end < 0:
    raise SystemExit("closing not found")
end += len("\n\t})\n")
replacement = "\n\t// JSON /api/* routes registered via registerAuthenticatedJSONAPIRoutes (see app_json_api.go).\n"
new_s = s[:start] + replacement + s[end:]
p.write_text(new_s, encoding="utf-8")
print("cut bytes", len(s) - len(new_s))
