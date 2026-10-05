"""
Gabungkan hasil export cookie extension + localStorage DevTools
menjadi format storage_state yang Playwright pahami (auth.json).

Pemakaian:
  python3 merge_auth.py cookies_export.json localstorage_export.json auth.json
"""
import json
import sys

def main():
    if len(sys.argv) != 4:
        print("Usage: python3 merge_auth.py <cookies.json> <localstorage.json> <output_auth.json>")
        sys.exit(1)

    cookies_file, localstorage_file, output_file = sys.argv[1], sys.argv[2], sys.argv[3]

    with open(cookies_file, "r", encoding="utf-8") as f:
        raw_cookies = json.load(f)

    with open(localstorage_file, "r", encoding="utf-8") as f:
        raw_localstorage = json.load(f)

    # Normalisasi cookie ke format Playwright (beberapa extension pakai field berbeda)
    cookies = []
    for c in raw_cookies:
        cookies.append({
            "name": c.get("name"),
            "value": c.get("value"),
            "domain": c.get("domain", "").lstrip("."),  # Playwright tidak suka leading dot di sebagian kasus
            "path": c.get("path", "/"),
            "expires": c.get("expirationDate", c.get("expires", -1)),
            "httpOnly": c.get("httpOnly", False),
            "secure": c.get("secure", True),
            "sameSite": c.get("sameSite", "Lax").capitalize() if c.get("sameSite") else "Lax",
        })

    storage_state = {
        "cookies": cookies,
        "origins": [
            {
                "origin": "https://emis.kemenag.go.id",
                "localStorage": raw_localstorage
            }
        ]
    }

    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(storage_state, f, ensure_ascii=False, indent=2)

    print(f"Berhasil, tersimpan di {output_file}")
    print(f"Total cookies: {len(cookies)}, total localStorage keys: {len(raw_localstorage)}")


if __name__ == "__main__":
    main()
