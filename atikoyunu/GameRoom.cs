namespace AtikOyunu
{
    using Microsoft.AspNetCore.SignalR;

    public record WasteDef(int Id, string Emoji, string Type, double X, double Speed);
    public record PlayerDto(string Name, int Score, int Correct, int Wrong, int Missed);

    public class Player
    {
        public string ConnectionId { get; init; } = "";
        public string Name { get; init; } = "";
        public int Score, Correct, Wrong, Missed;
        public HashSet<int> Done = new();   // aynı atık iki kez puanlanmasın
    }

    public class GameRoom
    {
        const int Countdown = 3, Duration = 60, WasteCount = 66;

        static readonly (string Emoji, string Type)[] Items =
        {
        ("🥤","plastic"), ("🧴","plastic"),
        ("📄","paper"),   ("📰","paper"),
        ("🍾","glass"),   ("🍷","glass"),
        ("🍎","organic"), ("🍌","organic"), ("🥕","organic")
    };

        readonly IHubContext<GameHub> _hub;
        readonly object _lock = new();
        readonly List<Player> _players = new();
        List<WasteDef> _wastes = new();
        bool _running;
        DateTime _deadline;

        public GameRoom(IHubContext<GameHub> hub) => _hub = hub;

        public async Task<string> Join(string connId, string name)
        {
            name = (name ?? "").Trim();
            if (name.Length == 0) return "İsim boş olamaz.";
            if (name.Length > 15) name = name[..15];

            lock (_lock)
            {
                if (_players.Any(p => p.ConnectionId == connId)) return "ok";
                if (_running) return "Oyun devam ediyor, bitmesini bekle.";
                if (_players.Any(p => p.Name.Equals(name, StringComparison.OrdinalIgnoreCase)))
                    return "Bu isim kullanılıyor.";

                _players.Add(new Player { ConnectionId = connId, Name = name });
            }

            await _hub.Groups.AddToGroupAsync(connId, "room");
            await BroadcastLobby();
            return "ok";
        }

        public async Task Leave(string connId)
        {
            lock (_lock) _players.RemoveAll(p => p.ConnectionId == connId);
            await BroadcastLobby();
            await BroadcastScores();
        }

        // Sadece host (listedeki ilk oyuncu) başlatabilir
        public async Task Start(string connId)
        {
            lock (_lock)
            {
                if (_running || _players.Count == 0 || _players[0].ConnectionId != connId)
                    return;

                _running = true;

                foreach (var p in _players)
                {
                    p.Score = p.Correct = p.Wrong = p.Missed = 0;
                    p.Done.Clear();
                }

                var rnd = new Random();
                _wastes = Enumerable.Range(0, WasteCount).Select(i =>
                {
                    var it = Items[rnd.Next(Items.Length)];
                    return new WasteDef(i, it.Emoji, it.Type,
                                        rnd.NextDouble(),
                                        1.5 + rnd.NextDouble() * 2.5);
                }).ToList();

                _deadline = DateTime.UtcNow.AddSeconds(Countdown + Duration + 1);
            }

            await _hub.Clients.Group("room")
                      .SendAsync("GameStarted", _wastes, Countdown, Duration);

            _ = EndAfterDelay();
        }

        public async Task Report(string connId, int wasteId, string? binType)
        {
            lock (_lock)
            {
                if (!_running || DateTime.UtcNow > _deadline) return;

                var p = _players.FirstOrDefault(x => x.ConnectionId == connId);
                if (p == null || wasteId < 0 || wasteId >= _wastes.Count) return;
                if (!p.Done.Add(wasteId)) return;   // zaten işlendi

                if (binType == null) { p.Missed++; p.Score -= 2; }
                else if (_wastes[wasteId].Type == binType) { p.Correct++; p.Score += 10; }
                else { p.Wrong++; p.Score -= 5; }
            }

            await BroadcastScores();
        }

        async Task EndAfterDelay()
        {
            await Task.Delay(TimeSpan.FromSeconds(Countdown + Duration + 2));

            List<PlayerDto> results;
            lock (_lock)
            {
                _running = false;
                results = Snapshot();
            }

            await _hub.Clients.Group("room").SendAsync("GameEnded", results);
            await BroadcastLobby();
        }

        async Task BroadcastScores()
        {
            List<PlayerDto> s;
            lock (_lock) s = Snapshot();
            await _hub.Clients.Group("room").SendAsync("ScoreBoard", s);
        }

        async Task BroadcastLobby()
        {
            string[] names; string host;
            lock (_lock)
            {
                names = _players.Select(p => p.Name).ToArray();
                host = _players.FirstOrDefault()?.Name ?? "";
            }
            await _hub.Clients.Group("room").SendAsync("LobbyUpdated", names, host);
        }

        List<PlayerDto> Snapshot() =>
            _players.OrderByDescending(p => p.Score)
                    .Select(p => new PlayerDto(p.Name, p.Score, p.Correct, p.Wrong, p.Missed))
                    .ToList();
    }
}
