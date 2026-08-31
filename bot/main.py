import discord
import aiohttp
import os
from dotenv import load_dotenv

load_dotenv()

DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)

async def send_to_backend(channel: str, content: str):
    async with aiohttp.ClientSession() as session:
        payload = {"channel": channel, "content": content}
        try:
            async with session.post(f"{BACKEND_URL}/events", json=payload) as resp:
                if resp.status == 201:
                    print(f"[OK] Event sent to backend: {content[:50]}")
                else:
                    print(f"[WARN] Backend responded with status {resp.status}")
        except Exception as e:
            print(f"[ERROR] Failed to reach backend: {e}")

@client.event
async def on_ready():
    print(f"[INFO] Meowmico is online as {client.user}!")

@client.event
async def on_message(message):
    if message.author == client.user:
        return

    print(f"[INFO] Message in #{message.channel.name}: {message.content[:50]}")
    await send_to_backend(message.channel.name, message.content)

client.run(DISCORD_TOKEN)