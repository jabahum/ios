"""Remove protected /api segments now covered by registerAuthenticatedJSONAPIRoutes (app_json_api.go)."""
from pathlib import Path

p = Path(__file__).resolve().parents[1] / "internal/routes/routes.go"
s = p.read_text(encoding="utf-8")

def cut_between(s: str, start_marker: str, end_marker: str) -> str:
    a = s.find(start_marker)
    if a < 0:
        raise SystemExit(f"missing start: {start_marker[:50]}")
    b = s.find(end_marker, a)
    if b < 0:
        raise SystemExit(f"missing end after start: {end_marker[:50]}")
    return s[:a] + s[b:]

# 1) Lab API (keep VHF CIF HTML below)
s = cut_between(
    s,
    "\t// Lab Sample Types API routes (protected - require authentication)\n",
    "\t// VHF CIF routes\n",
)

# 2) VHF vhf-cases (keep // Inventory items and HTML inventory)
a = s.find("\t// VHF API routes\n")
if a < 0:
    raise SystemExit("vhf-cases comment not found")
b = s.find("\t// Inventory items\n", a)
if b < 0:
    raise SystemExit("// Inventory items not found after vhf-cases")
s = s[:a] + s[b:]

# 3) Inventory JSON trio (keep HTML inventory below)
s = cut_between(
    s,
    "\t// Inventory API routes\n\tprotected.Get(\"/api/inventory/items\"",
    "\t// Reports routes - More specific routes must come first\n",
)

# 4) Comprehensive JSON API block
start = s.find("\t// ===== COMPREHENSIVE API ENDPOINTS")
if start < 0:
    raise SystemExit("comprehensive block not found")
end = s.find("\tprotected.Delete(\"/api/alerts/:id\", func(c *fiber.Ctx) error {", start)
if end < 0:
    raise SystemExit("alerts delete not found")
end = s.find("\n\t})\n", end)
if end < 0:
    raise SystemExit("alerts close not found")
end += len("\n\t})\n")
s = s[:start] + "\t// JSON /api/* routes: registerAuthenticatedJSONAPIRoutes (app_json_api.go).\n\n" + s[end:]

p.write_text(s, encoding="utf-8")
print("OK", p)
