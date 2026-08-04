#!/usr/bin/env python3
# ================================================================================
# VEILED PRIME X - HIGH-VELOCITY B2B LEAD HARVESTER ENGINE
# Path: /home/ronronalds/vpx_scraper.py
# ================================================================================

import sqlite3
import urllib.request
import urllib.parse
from html.parser import HTMLParser
import random
import time
import sys
import os
import concurrent.futures

USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2.1 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
]

class MicroParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.emails = set()
        self.links = set()

    def handle_starttag(self, tag, attrs):
        if tag == "a":
            for attr, value in attrs:
                if attr == "href":
                    if value.startswith("mailto:"):
                        email = value.replace("mailto:", "").split("?")[0].strip()
                        if "@" in email:
                            self.emails.add(email)
                    elif value.startswith("http"):
                        self.links.add(value)

    def handle_data(self, data):
        # Fallback text parsing regex simulation for text-wrapped emails
        words = data.split()
        for word in words:
            if "@" in word and "." in word:
                cleaned = word.strip("(),:;<>\"'")
                if len(cleaned) > 5 and len(cleaned) < 50:
                    self.emails.add(cleaned)

class LeadEngine:
    def __init__(self, db_path="./vpx_leads.db"):
        self.db_path = db_path
        self.init_storage()

    def init_storage(self):
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("PRAGMA journal_mode=WAL;")
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS b2b_leads (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                domain TEXT UNIQUE,
                email TEXT,
                status TEXT DEFAULT 'UNVERIFIED',
                harvested_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );
        """)
        conn.commit()
        conn.close()

    def fetch_target(self, target_url):
        req = urllib.request.Request(
            target_url,
            headers={"User-Agent": random.choice(USER_AGENTS)}
        )
        try:
            with urllib.request.urlopen(req, timeout=7) as response:
                if response.status == 200:
                    return response.read().decode("utf-8", errors="ignore")
        except Exception:
            return None
        return None

    def process_pipeline(self, target_url):
        domain = urllib.parse.urlparse(target_url).netloc
        if not domain:
            return f"[-] Failed: Invalid URL format -> {target_url}"

        html_content = self.fetch_target(target_url)
        if not html_content:
            return f"[-] Connection Timeout / Dropped: {domain}"

        parser = MicroParser()
        parser.feed(html_content)

        if parser.emails:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()
            for email in parser.emails:
                try:
                    cursor.execute(
                        "INSERT INTO b2b_leads (domain, email) VALUES (?, ?) ON CONFLICT(domain) DO NOTHING",
                        (domain, email)
                    )
                except sqlite3.Error:
                    pass
            conn.commit()
            conn.close()
            return f"[+] Extracted: {domain} -> Found {len(parser.emails)} Leads."
        
        return f"[*] Scanned: {domain} (No visible email matrices located)"

def execute_mesh_crawl(target_list):
    engine = LeadEngine()
    print(f"[*] Activating Lead Ingestion Matrix on {len(target_list)} core endpoints...")
    
    # Run high-concurrency multi-threaded networking pool
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
        results = list(executor.map(engine.process_pipeline, target_list))
        for res in results:
            print(res)

if __name__ == "__main__":
    # Test directory stack target targets
    seed_targets = [
        "https://ycombinator.com",
        "https://indiehackers.com",
        "https://remoteok.com"
    ]
    execute_mesh_crawl(seed_targets)
