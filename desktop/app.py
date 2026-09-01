import tkinter as tk
from PIL import Image, ImageTk, ImageSequence
import asyncio
import websockets
import json
import threading
import os
import time

BACKEND_WS = "ws://localhost:8080/ws"
ASSETS_DIR = os.path.join(os.path.dirname(__file__), "assets")

class MeowmicoApp:
    def __init__(self, root):
        self.root = root
        self.root.title("Meowmico")
        self.root.geometry("300x420")
        self.root.configure(bg="#16131F")
        self.root.resizable(False, False)

        # Keep window on top
        self.root.attributes("-topmost", True)

        # Current animation state
        self.current_anim = "idle"
        self.frames = {}
        self.frame_index = 0
        self.last_message_time = time.time()

        self._build_ui()
        self._load_animations()
        self._animate()
        self._start_ws_thread()
        self._check_idle()

    def _build_ui(self):
        # Title bar
        title = tk.Label(
            self.root, text="MEOWMICO",
            bg="#4A3F4B", fg="#F0D9E4",
            font=("Courier", 11, "bold"),
            pady=6
        )
        title.pack(fill=tk.X)

        # Cat canvas
        self.canvas = tk.Canvas(
            self.root, width=160, height=160,
            bg="#16131F", highlightthickness=0
        )
        self.canvas.pack(pady=10)
        self.cat_image = self.canvas.create_image(80, 80, anchor=tk.CENTER)

        # Speech bubble
        self.bubble = tk.Label(
            self.root,
            text="purrr... watching your servers~",
            bg="#4A3F4B", fg="#F0D9E4",
            font=("Courier", 9),
            wraplength=240,
            justify=tk.LEFT,
            padx=10, pady=8
        )
        self.bubble.pack(fill=tk.X, padx=16)

        # Events label
        self.events_label = tk.Label(
            self.root, text="0 events",
            bg="#16131F", fg="#8D6C79",
            font=("Courier", 8)
        )
        self.events_label.pack(pady=4)

        # Input row
        input_frame = tk.Frame(self.root, bg="#16131F")
        input_frame.pack(fill=tk.X, padx=16, pady=4)

        self.input_var = tk.StringVar()
        self.input_field = tk.Entry(
            input_frame,
            textvariable=self.input_var,
            bg="#4A3F4B", fg="#F0D9E4",
            insertbackground="#F0D9E4",
            font=("Courier", 9),
            relief=tk.FLAT,
            bd=4
        )
        self.input_field.pack(side=tk.LEFT, fill=tk.X, expand=True)
        self.input_field.bind("<Return>", self._on_send)

        send_btn = tk.Button(
            input_frame, text="send",
            bg="#8D6C79", fg="#F0D9E4",
            font=("Courier", 9),
            relief=tk.FLAT,
            cursor="hand2",
            command=self._on_send
        )
        send_btn.pack(side=tk.RIGHT, padx=(6, 0))

        # Status
        self.status_label = tk.Label(
            self.root, text="connecting...",
            bg="#16131F", fg="#C1A0AC",
            font=("Courier", 8)
        )
        self.status_label.pack(pady=4)

        self.event_count = 0

    def _load_animations(self):
        anim_names = ["idle", "happy", "run", "sit", "jump"]
        for name in anim_names:
            path = os.path.join(ASSETS_DIR, f"{name}.gif")
            if os.path.exists(path):
                gif = Image.open(path)
                frames = []
                try:
                    while True:
                        frame = gif.copy().convert("RGBA")
                        frame = frame.resize((100, 100), Image.NEAREST)
                        frames.append(ImageTk.PhotoImage(frame))
                        gif.seek(gif.tell() + 1)
                except EOFError:
                    pass
                self.frames[name] = frames

    def _animate(self):
        frames = self.frames.get(self.current_anim, self.frames.get("idle", []))
        if frames:
            self.frame_index = (self.frame_index + 1) % len(frames)
            self.canvas.itemconfig(self.cat_image, image=frames[self.frame_index])
        self.root.after(100, self._animate)

    def set_animation(self, name, message=None):
        if name in self.frames:
            self.current_anim = name
            self.frame_index = 0
        if message:
            self.bubble.config(text=message)

    def _on_send(self, event=None):
        text = self.input_var.get().strip()
        if not text:
            return
        self.input_var.set("")
        self.set_animation("happy", f"meow! you said: {text[:40]}")
        self.root.after(3000, lambda: self.set_animation("idle"))

    def _on_discord_event(self, data):
        self.event_count += 1
        self.events_label.config(text=f"{self.event_count} events")
        channel = data.get("channel", "unknown")
        content = data.get("content", "")[:40]
        self.set_animation("run", f"new message in #{channel}!\n{content}")
        self.last_message_time = time.time()
        self.root.after(4000, lambda: self.set_animation("idle"))

    def _check_idle(self):
        if time.time() - self.last_message_time > 30:
            if self.current_anim not in ("happy", "run", "jump"):
                self.set_animation("sit", "zzz... all quiet~")
        self.root.after(5000, self._check_idle)

    def _start_ws_thread(self):
        thread = threading.Thread(target=self._ws_loop, daemon=True)
        thread.start()

    def _ws_loop(self):
        async def connect():
            while True:
                try:
                    async with websockets.connect(BACKEND_WS) as ws:
                        self.root.after(0, lambda: self.status_label.config(
                            text="connected", fg="#a8c5a0"
                        ))
                        self.root.after(0, lambda: self.set_animation(
                            "jump", "backend connected! ready~"
                        ))
                        async for message in ws:
                            try:
                                data = json.loads(message)
                                if data.get("content"):
                                    self.root.after(0, lambda d=data: self._on_discord_event(d))
                            except json.JSONDecodeError:
                                pass
                except Exception:
                    self.root.after(0, lambda: self.status_label.config(
                        text="reconnecting...", fg="#C1A0AC"
                    ))
                    await asyncio.sleep(5)

        asyncio.run(connect())

if __name__ == "__main__":
    root = tk.Tk()
    app = MeowmicoApp(root)
    root.mainloop()