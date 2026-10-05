const EMIS_HOST = "emis.kemenag.go.id";

const urlInput = document.getElementById("coreAppUrl");
const modeSelect = document.getElementById("mode");
const runBtn = document.getElementById("runBtn");
const statusBox = document.getElementById("status");
const logBox = document.getElementById("log");
const usernameInput = document.getElementById("adminUsername");
const passwordInput = document.getElementById("adminPassword");
const saveCredBtn = document.getElementById("saveCredBtn");
const credDetails = document.getElementById("credDetails");

let pollTimer = null;

// muat pengaturan tersimpan
chrome.storage.local.get(["coreAppUrl", "mode", "adminUsername", "adminPassword"], (data) => {
  if (data.coreAppUrl) urlInput.value = data.coreAppUrl;
  if (data.mode) modeSelect.value = data.mode;
  if (data.adminUsername) usernameInput.value = data.adminUsername;
  if (data.adminPassword) passwordInput.value = data.adminPassword;

  // kalau kredensial sudah ada, collapse bagian itu biar popup ringkas
  if (data.adminUsername && data.adminPassword) {
    credDetails.removeAttribute("open");
  } else {
    credDetails.setAttribute("open", "true");
  }
});

[urlInput, modeSelect].forEach((el) => {
  el.addEventListener("change", () => {
    chrome.storage.local.set({ coreAppUrl: urlInput.value, mode: modeSelect.value });
  });
});

saveCredBtn.addEventListener("click", () => {
  chrome.storage.local.set(
    { adminUsername: usernameInput.value, adminPassword: passwordInput.value },
    () => showStatus("ok", "Kredensial tersimpan di extension ini.")
  );
});

function showStatus(type, msg) {
  statusBox.className = type;
  statusBox.textContent = msg;
  statusBox.style.display = "block";
}

async function getActiveEmisTab() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab || !tab.url || !tab.url.includes(EMIS_HOST)) {
    throw new Error("Buka dan login ke emis.kemenag.go.id dulu, lalu klik lagi di tab itu.");
  }
  return tab;
}

async function getCookies() {
  const a = await chrome.cookies.getAll({ domain: "emis.kemenag.go.id" });
  const b = await chrome.cookies.getAll({ domain: "kemenag.go.id" });
  const merged = [...a, ...b];
  const seen = new Set();
  const unique = [];
  for (const c of merged) {
    const key = `${c.domain}|${c.name}|${c.path}`;
    if (!seen.has(key)) {
      seen.add(key);
      unique.push(c);
    }
  }
  return unique.map((c) => ({
    name: c.name,
    value: c.value,
    domain: c.domain,
    path: c.path,
    expirationDate: c.expirationDate || -1,
    httpOnly: c.httpOnly,
    secure: c.secure,
    sameSite: c.sameSite,
  }));
}

async function getLocalStorage(tabId) {
  const [{ result }] = await chrome.scripting.executeScript({
    target: { tabId },
    func: () => Object.keys(localStorage).map((k) => ({ name: k, value: localStorage.getItem(k) })),
  });
  return result;
}

async function loginAndGetToken(coreAppUrl) {
  const username = usernameInput.value;
  const password = passwordInput.value;

  if (!username || !password) {
    throw new Error("Isi dan simpan kredensial admin dulu (klik 'Kredensial Admin' di bawah).");
  }

  const res = await fetch(`${coreAppUrl}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || "Login admin gagal. Cek username/password.");
  }

  return data.token;
}

async function pollStatus(coreAppUrl, token) {
  try {
    const res = await fetch(`${coreAppUrl}/api/v1/admin/scraper/emis/status`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    logBox.style.display = "block";
    logBox.textContent = (data.log_tail || []).join("\n");
    logBox.scrollTop = logBox.scrollHeight;

    if (!data.running) {
      clearInterval(pollTimer);
      showStatus("ok", "Scraping selesai.");
      runBtn.disabled = false;
      runBtn.textContent = "Ambil Sesi & Mulai Scraping";
    }
  } catch (e) {
    // diamkan, coba lagi di interval berikutnya
  }
}

runBtn.addEventListener("click", async () => {
  statusBox.style.display = "none";
  logBox.style.display = "none";
  runBtn.disabled = true;
  runBtn.textContent = "Login admin...";

  const coreAppUrl = urlInput.value.replace(/\/$/, "");
  const mode = modeSelect.value;

  if (!coreAppUrl) {
    showStatus("err", "Isi URL core-app dulu.");
    runBtn.disabled = false;
    runBtn.textContent = "Ambil Sesi & Mulai Scraping";
    return;
  }

  try {
    const token = await loginAndGetToken(coreAppUrl);

    runBtn.textContent = "Mengambil sesi EMIS...";
    const tab = await getActiveEmisTab();

    const [cookies, localStorageData] = await Promise.all([getCookies(), getLocalStorage(tab.id)]);

    if (cookies.length === 0 || localStorageData.length === 0) {
      throw new Error("Sesi EMIS kosong, pastikan sudah login penuh (bukan halaman login).");
    }

    runBtn.textContent = "Mengirim ke core-app...";

    const res = await fetch(`${coreAppUrl}/api/v1/admin/scraper/emis/trigger-json?mode=${mode}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ cookies, localStorage: localStorageData }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || "Gagal memulai scraping.");
    }

    showStatus("ok", `Scraping dimulai (mode: ${data.mode}). Memantau progres...`);
    runBtn.textContent = "Sedang berjalan...";
    pollTimer = setInterval(() => pollStatus(coreAppUrl, token), 3000);
  } catch (err) {
    showStatus("err", err.message);
    runBtn.disabled = false;
    runBtn.textContent = "Ambil Sesi & Mulai Scraping";
  }
});
