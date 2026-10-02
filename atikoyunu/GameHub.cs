namespace AtikOyunu
{
    using Microsoft.AspNetCore.SignalR;
    public class GameHub : Hub
    {
        private readonly GameRoom _room;
        public GameHub(GameRoom room) => _room = room;
        public Task<string> Join(string name) =>
            _room.Join(Context.ConnectionId, name);
        public Task StartGame() =>
            _room.Start(Context.ConnectionId);

        // binType null ise atık kaçmış demektir
        public Task Report(int wasteId, string? binType) =>
            _room.Report(Context.ConnectionId, wasteId, binType);

        public override async Task OnDisconnectedAsync(Exception? ex)
        {
            await _room.Leave(Context.ConnectionId);
            await base.OnDisconnectedAsync(ex);
        }
    }
}