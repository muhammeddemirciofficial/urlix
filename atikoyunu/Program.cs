using AtikOyunu;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddSignalR();
builder.Services.AddSingleton<GameRoom>();

var app = builder.Build();

app.UseDefaultFiles();   // / adresinde index.html açılır
app.UseStaticFiles();
app.MapHub<GameHub>("/gamehub");

app.Run();