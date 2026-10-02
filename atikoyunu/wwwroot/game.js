/* =====================================================
   ATIK TOPLAMA OYUNU - İSTEMCİ (SignalR ile çok oyunculu)
   Atık listesi, puanlama ve süre sunucudan yönetilir.
   ===================================================== */

/* ---------- DOM ELEMANLARI ---------- */

const gameArea        = document.getElementById("gameArea");
const scoreElement    = document.getElementById("score");
const timeElement     = document.getElementById("time");
const correctElement  = document.getElementById("correct");

const startScreen     = document.getElementById("startScreen");
const lobbyScreen     = document.getElementById("lobbyScreen");
const gameOverScreen  = document.getElementById("gameOverScreen");

const playerNameInput = document.getElementById("playerName");
const lobbyList       = document.getElementById("lobbyList");
const lobbyInfo       = document.getElementById("lobbyInfo");
const hostStartBtn    = document.getElementById("hostStartButton");
const countdownEl     = document.getElementById("countdown");
const liveBoard       = document.getElementById("liveBoard");

const finalScore      = document.getElementById("finalScore");
const finalCorrect    = document.getElementById("finalCorrect");
const finalWrong      = document.getElementById("finalWrong");
const finalMissed     = document.getElementById("finalMissed");
const carbonSaved     = document.getElementById("carbonSaved");
const carbonReduction = document.getElementById("carbonReduction");
const leaderboardList = document.getElementById("leaderboardList");

/* ---------- KARBON KATSAYILARI (kg CO2e, örnek tahmin) ---------- */

const carbonFactors = {
    plastic: 0.08,
    paper: 0.04,
    glass: 0.12,
    organic: 0.03
};

/* ---------- OYUN DEĞİŞKENLERİ ---------- */

let score = 0;
let correct = 0;
let wrong = 0;
let missed = 0;
let time = 60;

let gameRunning = false;
let gameTimer = null;
let spawnTimeouts = [];
let collectedWastes = [];

let myName = "";
let leaving = false;   // Çıkış yaparken "bağlantı koptu" uyarısı çıkmasın

/* ---------- SIGNALR BAĞLANTISI ---------- */

const connection = new signalR.HubConnectionBuilder()
    .withUrl("/gamehub")
    .build();

connection.onclose(() => {
    if (leaving) return;
    alert("Sunucu bağlantısı koptu. Sayfa yenileniyor.");
    location.reload();
});

/* ---------- SUNUCUDAN GELEN OLAYLAR ---------- */

connection.on("LobbyUpdated", (players, hostName) => {

    lobbyList.innerHTML = "";

    players.forEach(n => {
        const li = document.createElement("li");
        li.textContent =
            (n === hostName ? "👑 " : "👤 ") + n + (n === myName ? " (sen)" : "");
        lobbyList.appendChild(li);
    });

    const isHost = hostName === myName;

    hostStartBtn.style.display = isHost ? "inline-block" : "none";
    lobbyInfo.textContent = isHost
        ? "Herkes hazır olunca oyunu başlat."
        : hostName + " oyunu başlatınca oyun başlayacak...";
});

connection.on("GameStarted", (wastes, countdown, duration) => {

    // Önceki oyundan kalanları temizle
    document.querySelectorAll(".waste").forEach(w => w.remove());
    spawnTimeouts.forEach(clearTimeout);
    spawnTimeouts = [];
    clearInterval(gameTimer);

    score = correct = wrong = missed = 0;
    collectedWastes = [];
    time = duration;
    updateUI();

    startScreen.style.display = "none";
    lobbyScreen.style.display = "none";
    gameOverScreen.style.display = "none";
    liveBoard.innerHTML = "";
    liveBoard.style.display = "block";

    runCountdown(countdown, () => {

        gameRunning = true;

        gameTimer = setInterval(() => {
            time--;
            timeElement.textContent = Math.max(time, 0);

            if (time <= 0) {
                clearInterval(gameTimer);
                gameRunning = false;   // Sonuç ekranını sunucu gönderecek
            }
        }, 1000);

        wastes.forEach(def => {
            spawnTimeouts.push(
                setTimeout(() => spawnWaste(def), def.id * 900)
            );
        });
    });
});

connection.on("ScoreBoard", list => {

    // Kendi değerlerimizi sunucudan al
    const me = list.find(p => p.name === myName);

    if (me) {
        score = me.score;
        correct = me.correct;
        wrong = me.wrong;
        missed = me.missed;
        updateUI();
    }

    // Canlı ilk 5
    liveBoard.innerHTML = "";

    list.slice(0, 5).forEach((p, i) => {
        const row = document.createElement("div");
        row.textContent = (i + 1) + ". " + p.name + ": " + p.score;
        if (p.name === myName) row.style.fontWeight = "bold";
        liveBoard.appendChild(row);
    });
});

connection.on("GameEnded", results => endGame(results));

/* ---------- GİRİŞ / LOBİ ---------- */

async function joinLobby() {

    const name = playerNameInput.value.trim();

    if (!name) {
        playerNameInput.placeholder = "Önce adını yaz!";
        playerNameInput.focus();
        return;
    }

    try {
        if (connection.state === signalR.HubConnectionState.Disconnected) {
            await connection.start();
        }

        myName = name;
        const result = await connection.invoke("Join", name);

        if (result !== "ok") {
            myName = "";
            playerNameInput.value = "";
            playerNameInput.placeholder = result;
            return;
        }

        startScreen.style.display = "none";
        lobbyScreen.style.display = "flex";

    } catch (e) {
        playerNameInput.value = "";
        playerNameInput.placeholder = "Sunucuya bağlanılamadı";
    }
}

document.getElementById("startButton").addEventListener("click", joinLobby);

playerNameInput.addEventListener("keydown", e => {
    if (e.key === "Enter") joinLobby();
});

hostStartBtn.addEventListener("click", () => {
    connection.invoke("StartGame").catch(() => {});
});

// Sonuç ekranından lobiye dön
document.getElementById("restartButton").addEventListener("click", () => {
    gameOverScreen.style.display = "none";
    lobbyScreen.style.display = "flex";
});

// Çıkış: bağlantıyı kapat, isim ekranına dön
document.getElementById("newPlayerButton").addEventListener("click", async () => {
    leaving = true;
    try { await connection.stop(); } catch (e) {}
    location.reload();
});

/* ---------- GERİ SAYIM ---------- */

function runCountdown(sec, done) {

    countdownEl.style.display = "flex";

    const tick = () => {
        if (sec <= 0) {
            countdownEl.style.display = "none";
            done();
            return;
        }
        countdownEl.textContent = sec--;
        setTimeout(tick, 1000);
    };

    tick();
}

/* ---------- ATIK OLUŞTUR (sunucudan gelen tanımla) ---------- */

function spawnWaste(def) {

    if (!gameRunning) return;

    const waste = document.createElement("div");

    waste.classList.add("waste");
    waste.textContent = def.emoji;
    waste.dataset.type = def.type;
    waste.dataset.id = def.id;

    waste.style.left = (def.x * (gameArea.clientWidth - 70)) + "px";
    waste.style.top = "-70px";

    gameArea.appendChild(waste);
    makeDraggable(waste);

    const fall = setInterval(() => {

        if (!gameRunning || !waste.parentElement) {
            clearInterval(fall);
            return;
        }

        // Sürüklenirken otomatik düşme yok
        if (waste.dataset.dragging === "true") return;

        // Güncel konumu DOM'dan oku (sürükleme sonrası zıplamayı önler)
        let y = parseFloat(waste.style.top) || 0;
        y += def.speed;
        waste.style.top = y + "px";

        // Kutuların üst kenarına ulaştıysa kaçırılmış sayılır
        if (y > getFloorY()) {
            clearInterval(fall);

            collectedWastes.push({
                type: waste.dataset.type,
                correct: false,
                missed: true
            });

            waste.remove();
            connection.invoke("Report", def.id, null).catch(() => {});
        }

    }, 16);
}

/* ---------- SÜRÜKLE-BIRAK ---------- */

function makeDraggable(element) {

    let offsetX = 0;
    let offsetY = 0;

    element.addEventListener("mousedown", startDrag);
    element.addEventListener("touchstart", startDrag, { passive: false });

    function getPoint(event) {
        if (event.touches && event.touches.length > 0) {
            return { x: event.touches[0].clientX, y: event.touches[0].clientY };
        }
        return { x: event.clientX, y: event.clientY };
    }

    function startDrag(event) {

        if (!gameRunning) return;

        event.preventDefault();

        element.dataset.dragging = "true";

        const rect = element.getBoundingClientRect();
        const p = getPoint(event);

        offsetX = p.x - rect.left;
        offsetY = p.y - rect.top;

        document.addEventListener("mousemove", drag);
        document.addEventListener("mouseup", stopDrag);
        document.addEventListener("touchmove", drag, { passive: false });
        document.addEventListener("touchend", stopDrag);
    }

    function drag(event) {

        event.preventDefault();

        const p = getPoint(event);
        const areaRect = gameArea.getBoundingClientRect();

        let x = p.x - areaRect.left - offsetX;
        let y = p.y - areaRect.top - offsetY;

        // Ekranın dışına çıkmasını engelle
        x = Math.max(0, Math.min(x, gameArea.clientWidth - 65));
        y = Math.max(0, Math.min(y, gameArea.clientHeight - 65));

        element.style.left = x + "px";
        element.style.top = y + "px";
    }

    function stopDrag() {

        element.dataset.dragging = "false";

        document.removeEventListener("mousemove", drag);
        document.removeEventListener("mouseup", stopDrag);
        document.removeEventListener("touchmove", drag);
        document.removeEventListener("touchend", stopDrag);

        checkBinCollision(element);
    }
}

/* ---------- KUTUYA BIRAKILDI MI? ---------- */

function checkBinCollision(waste) {

    if (!waste.parentElement) return;

    const w = waste.getBoundingClientRect();
    const cx = w.left + w.width / 2;
    const cy = w.top + w.height / 2;

    for (const bin of document.querySelectorAll(".bin")) {

        const b = bin.getBoundingClientRect();

        const inside =
            cx > b.left && cx < b.right &&
            cy > b.top  && cy < b.bottom;

        if (inside) {

            const wasteType = waste.dataset.type;
            const binType = bin.dataset.type;

            // Karbon hesabı için kaydet
            collectedWastes.push({
                type: wasteType,
                correct: wasteType === binType
            });

            // Puanı sunucu hesaplar
            connection
                .invoke("Report", Number(waste.dataset.id), binType)
                .catch(() => {});

            waste.remove();
            return;
        }
    }

    // Kutuya bırakılmadı: tekrar düşmeye devam etsin
    waste.dataset.dragging = "false";
}

/* ---------- YARDIMCI FONKSİYONLAR ---------- */

function updateUI() {
    scoreElement.textContent = score;
    correctElement.textContent = correct;
    timeElement.textContent = Math.max(time, 0);
}

function getFloorY() {
    const areaRect = gameArea.getBoundingClientRect();
    const binTop = document.querySelector(".bin").getBoundingClientRect().top;

    return binTop - areaRect.top - 40;
}

/* ---------- OYUNU BİTİR (sonuçlar sunucudan gelir) ---------- */

function endGame(results) {

    gameRunning = false;

    clearInterval(gameTimer);
    spawnTimeouts.forEach(clearTimeout);
    spawnTimeouts = [];

    document.querySelectorAll(".waste").forEach(w => w.remove());
    liveBoard.style.display = "none";
    countdownEl.style.display = "none";

    // Kesin değerler sunucudan
    const me = results.find(p => p.name === myName);

    if (me) {
        score = me.score;
        correct = me.correct;
        wrong = me.wrong;
        missed = me.missed;
    }

    /* Karbon hesabı */

    let savedCarbon = 0;   // Doğru ayrıştırılanlardan kazanılan
    let totalCarbon = 0;   // Karşılaşılan tüm atıkların potansiyeli

    collectedWastes.forEach(item => {
        const factor = carbonFactors[item.type] || 0;
        totalCarbon += factor;
        if (item.correct) savedCarbon += factor;
    });

    const percent = totalCarbon > 0
        ? Math.round((savedCarbon / totalCarbon) * 100)
        : 0;

    /* Sonuç ekranı */

    finalScore.textContent = "Skor: " + score;
    finalCorrect.textContent = correct;
    finalWrong.textContent = wrong;
    finalMissed.textContent = missed;

    carbonSaved.textContent = savedCarbon.toFixed(2);
    carbonReduction.textContent = "%" + percent;

    /* Sıralama (sunucudan gelen, tüm oyuncular için aynı) */

    leaderboardList.innerHTML = "";

    const medals = ["🥇", "🥈", "🥉"];

    results.forEach((p, i) => {

        const li = document.createElement("li");

        if (p.name === myName) li.classList.add("current");

        li.textContent = (medals[i] || (i + 1) + ".") + " " + p.name;

        const pts = document.createElement("span");
        pts.textContent = p.score;
        li.appendChild(pts);

        leaderboardList.appendChild(li);
    });

    gameOverScreen.style.display = "flex";
}
