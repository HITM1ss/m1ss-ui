#!/usr/bin/env python3
"""Exercise a disposable M1ss-UI instance; never point this at a live panel."""

import base64
import http.cookiejar
import json
import sys
import time
import urllib.parse
import urllib.request


panel, subscription = sys.argv[1:3]
opener = urllib.request.build_opener(
    urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
)


def request(path, form=None):
    data = urllib.parse.urlencode(form).encode() if form is not None else None
    req = urllib.request.Request(panel + path, data=data, headers={
        "Origin": panel,
        "X-Requested-With": "XMLHttpRequest",
    })
    with opener.open(req, timeout=10) as response:
        return response.read()


def api(path, form=None):
    result = json.loads(request("/app/api/" + path, form))
    assert result["success"], result
    return result["obj"]


def save(obj, data):
    return api("save", {"object": obj, "action": "new", "data": json.dumps(data)})


for attempt in range(60):
    try:
        assert b"M1ss-UI" in request("/app/")
        break
    except Exception:
        if attempt == 59:
            raise
        time.sleep(1)

api("login", {"user": "admin", "pass": "admin"})
assert api("settings")["subPort"] == "2096"
assert api("users"), "admin table not available"

clients = api("clients")["clients"] or []
if not any(client["name"] == "minimal-smoke" for client in clients):
    keys = api("keypairs?k=reality")
    private_key = next(line.split(":", 1)[1].strip() for line in keys if line.startswith("PrivateKey:"))
    public_key = next(line.split(":", 1)[1].strip() for line in keys if line.startswith("PublicKey:"))
    tls_config = {
        "name": "minimal-reality",
        "server": {"enabled": True, "server_name": "example.com", "reality": {
            "enabled": True, "private_key": private_key, "short_id": ["0123456789abcdef"],
            "handshake": {"server": "example.com", "server_port": 443},
        }},
        "client": {"enabled": True, "server_name": "example.com", "reality": {
            "enabled": True, "public_key": public_key,
        }, "utls": {"enabled": True, "fingerprint": "chrome"}},
    }
    tls_rows = save("tls", tls_config)["tls"]
    tls_id = next(row["id"] for row in tls_rows if row["name"] == "minimal-reality")
    inbounds = save("inbounds", {
        "type": "vless", "tag": "minimal-vless", "listen": "::", "listen_port": 24443,
        "tls_id": tls_id, "transport": {}, "out_json": {}, "addrs": [],
    })["inbounds"]
    inbound_id = next(row["id"] for row in inbounds if row["tag"] == "minimal-vless")
    save("clients", {
        "name": "minimal-smoke", "enable": True, "inbounds": [inbound_id],
        "config": {"vless": {"name": "minimal-smoke", "uuid": "11111111-1111-1111-1111-111111111111", "flow": "xtls-rprx-vision"}},
        "links": [], "volume": 0, "expiry": 0, "up": 0, "down": 0,
    })

for suffix in ("", "?format=json", "?format=clash"):
    with urllib.request.urlopen(subscription + "/sub/minimal-smoke" + suffix, timeout=10) as response:
        body = response.read()
    if not suffix:
        body = base64.b64decode(body)
    text = body.decode()
    assert "vless" in text.lower() and "xtls-rprx-vision" in text, text
    assert "reality" in text and "chrome" in text, text

# Survive several watchdog/statistics cycles and repeated authenticated polls.
for _ in range(9):
    status = api("status?r=sbd")["sbd"]
    assert status["running"], status
    api("load")
    time.sleep(5)
print("PASS: UI, SQLite, admin login, VLESS/Reality/Vision config, client management, three subscription formats and 45 seconds of polling")
