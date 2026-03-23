"""Extract protected /api routes from routes.go and emit registerAuthenticatedJSONAPIRoutes body."""
import re
from pathlib import Path

root = Path(__file__).resolve().parents[1]
routes = (root / "internal/routes/routes.go").read_text(encoding="utf-8")

lab_start = routes.find("\t// Lab Sample Types API routes (protected")
lab_end = routes.find("\n\t// VHF CIF routes", lab_start)
lab = routes[lab_start:lab_end]

vc_start = routes.find('\t// VHF API routes\n\tprotected.Get("/api/vhf-cases"', lab_end)
vc_end = routes.find("\n\n\t// Inventory items", vc_start)
vc = routes[vc_start:vc_end]

inv_start = routes.find('\t// Inventory API routes\n\tprotected.Get("/api/inventory/items"', vc_end)
inv_end = routes.find("\n\t// Reports routes - More specific", inv_start)
inv = routes[inv_start:inv_end]

comp_start = routes.find("\t// ===== COMPREHENSIVE API ENDPOINTS", inv_end)
comp_end = routes.find("\n}\n\n// RouteSurveillance", comp_start)
comp = routes[comp_start:comp_end]


def transform(block: str) -> str:
    block = block.replace("protected.Get(", "app.Get(")
    block = block.replace("protected.Post(", "app.Post(")
    block = block.replace("protected.Put(", "app.Put(")
    block = block.replace("protected.Delete(", "app.Delete(")
    return block


body = "\n".join([transform(lab), transform(vc), transform(inv), transform(comp)])

body = re.sub(
    r'(\n\tapp\.(?:Get|Post|Put|Delete)\("[^"]+",)\s*',
    r"\1 auth, ",
    body,
)

# Drop route blocks whose registration line matches (exact path string).
skip_paths = {
    'app.Get("/api/users",',
    'app.Post("/api/cases",',
    'app.Get("/api/facilities",',
    'app.Get("/api/outbreaks",',
}

lines = body.split("\n")
out_lines = []
i = 0
while i < len(lines):
    line = lines[i]
    skip = False
    for sp in skip_paths:
        if sp in line:
            skip = True
            break
    if skip:
        depth = line.count("{") - line.count("}")
        i += 1
        while i < len(lines) and depth > 0:
            depth += lines[i].count("{") - lines[i].count("}")
            i += 1
        continue
    out_lines.append(line)
    i += 1

body = "\n".join(out_lines)

out = root / "internal/routes/app_json_api_routes_body.go.txt"
out.write_text(body, encoding="utf-8")
print("Wrote", out, "lines", body.count("\n"))
