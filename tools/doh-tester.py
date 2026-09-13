#!/usr/bin/env python3
"""
ONCIC Network Monitoring - DoH Server Tester
DNS over HTTPS (DoH) endpoint testing and domain resolution validator

Tests 100+ DoH servers worldwide and validates domain resolution.
Part of ONCIC Network Monitoring Platform / Created by Medo
"""

import argparse
import asyncio
import json
import random
import struct
import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple

import httpx
from rich.console import Console
from rich.progress import (
    BarColumn,
    Progress,
    SpinnerColumn,
    TextColumn,
    TimeElapsedColumn,
)
from rich.rule import Rule
from rich.table import Table
from rich import box

# ──────────────────────────────────────────────────────────────────────────────
# CONFIGURATION
# ──────────────────────────────────────────────────────────────────────────────

DOMAINS_TO_RESOLVE = [
    "youtube.com",
    "whatsapp.com",
    "x.com",
    "instagram.com",
    "telegram.org",
    "facebook.com",
]

TIMEOUT_SECONDS = 8.0
CONCURRENCY_LIMIT = 25
PROBE_DOMAIN = "example.com"

# (provider name, DoH URL, region / label)
DOH_SERVERS: List[Tuple[str, str, str]] = [
    ("aaflalo.me", "https://dns-nyc.aaflalo.me/dns-query", "US"),
    ("AdGuard", "https://dns.adguard.com/dns-query", "Default"),
    ("AdGuard Family", "https://dns-family.adguard.com/dns-query", "Family"),
    ("Adhole", "https://uk.adhole.org/dns-query", "United Kingdom"),
    ("Adhole", "https://de.adhole.org/dns-query", "Germany"),
    ("Adhole", "https://sg.adhole.org/dns-query", "Singapore"),
    ("Adhole", "https://us-central.adhole.org/dns-query", "US Central"),
    ("Adhole", "https://us-east.adhole.org/dns-query", "US East"),
    ("AhaDNS", "https://doh.nl.ahadns.net/dns-query", "Netherlands"),
    ("AhaDNS", "https://doh.in.ahadns.net/dns-query", "India"),
    ("AhaDNS", "https://doh.la.ahadns.net/dns-query", "Los Angeles"),
    ("AhaDNS", "https://doh.ny.ahadns.net/dns-query", "New York"),
    ("AhaDNS", "https://doh.pl.ahadns.net/dns-query", "Poland"),
    ("AhaDNS", "https://doh.it.ahadns.net/dns-query", "Italy"),
    ("AhaDNS", "https://doh.es.ahadns.net/dns-query", "Spain"),
    ("AhaDNS", "https://doh.no.ahadns.net/dns-query", "Norway"),
    ("AhaDNS", "https://doh.chi.ahadns.net/dns-query", "Chicago"),
    ("AhaDNS", "https://doh.au.ahadns.net/dns-query", "Australia"),
    ("Alibaba Public DNS", "https://dns.alidns.com/dns-query", "-"),
    ("Andrews & Arnold", "https://dns.aa.net.uk/dns-query", "-"),
    ("alekberg", "https://dnses.alekberg.net/dns-query", "Spain"),
    ("alekberg", "https://dnsnl.alekberg.net/dns-query", "Holland"),
    ("alekberg", "https://dnsse.alekberg.net/dns-query", "Sweden"),
    ("Arapuyaril", "https://dns.arapurayil.com/dns-query", "-"),
    ("Association 42l", "https://doh.42l.fr/dns-query", "-"),
    ("BebasDNS", "https://dns.doh.my.id/dns-query", "Singapore"),
    ("blahdns.com", "https://doh-ch.blahdns.com/dns-query", "Switzerland"),
    ("blahdns.com", "https://doh-sg.blahdns.com/dns-query", "Singapore"),
    ("blahdns.com", "https://doh-fi.blahdns.com/dns-query", "Finland"),
    ("blahdns.com", "https://doh-jp.blahdns.com/dns-query", "Japan"),
    ("blahdns.com", "https://doh-de.blahdns.com/dns-query", "Germany"),
    ("Blokada DNS", "https://dns.blokada.org/dns-query", "-"),
    ("Charter/Spectrum", "https://doh-01.spectrum.com/dns-query", "California"),
    ("Charter/Spectrum", "https://doh-02.spectrum.com/dns-query", "Texas"),
    ("CIRA Canadian Shield", "https://private.canadianshield.cira.ca/dns-query", "Private"),
    ("CIRA Canadian Shield", "https://protected.canadianshield.cira.ca/dns-query", "Protected"),
    ("CIRA Canadian Shield", "https://family.canadianshield.cira.ca/dns-query", "Family"),
    ("Cisco Umbrella (OpenDNS)", "https://doh.opendns.com/dns-query", "Standard"),
    ("Cisco Umbrella (OpenDNS)", "https://doh.familyshield.opendns.com/dns-query", "FamilyShield"),
    ("CleanBrowsing", "https://doh.cleanbrowsing.org/doh/family-filter/", "Family Filter"),
    ("Cloudflare", "https://cloudflare-dns.com/dns-query", "Main"),
    ("Cloudflare", "https://mozilla.cloudflare-dns.com/dns-query", "Mozilla"),
    ("Cloudflare", "https://security.cloudflare-dns.com/dns-query", "Block Malware"),
    ("Cloudflare", "https://family.cloudflare-dns.com/dns-query", "Block Malware+Adult"),
    ("Cloudflare", "https://dns64.cloudflare-dns.com/dns-query", "DNS64"),
    ("Comcast", "https://doh.xfinity.com/dns-query", "-"),
    ("ControlD", "https://freedns.controld.com/p0", "Unfiltered"),
    ("ControlD", "https://freedns.controld.com/p1", "Block Malware"),
    ("ControlD", "https://freedns.controld.com/p2", "Block Malware+Ads"),
    ("ControlD", "https://freedns.controld.com/p3", "Block Malware+Ads+Social"),
    ("CZ.NIC", "https://odvr.nic.cz/dns-query", "-"),
    ("Digitale Gesellschaft", "https://dns.digitale-gesellschaft.ch/dns-query", "-"),
    ("dns.flatuslifir.is", "https://dns.flatuslifir.is/dns-query", "-"),
    ("DNS.SB", "https://doh.dns.sb/dns-query", "-"),
    ("dnsforge.de", "https://dnsforge.de/dns-query", "-"),
    ("dnsHome.de", "https://dns.dnshome.de/dns-query", "-"),
    ("DNSlify", "https://doh.dnslify.com/dns-query", "-"),
    ("doh.li", "https://doh.li/dns-query", "-"),
    ("dnswarden", "https://doh.asia.dnswarden.com/adblock", "Singapore/Adblock"),
    ("dnswarden", "https://doh.asia.dnswarden.com/uncensored", "Singapore/Uncensored"),
    ("dnswarden", "https://doh.asia.dnswarden.com/adultfilter", "Singapore/AdultFilter"),
    ("dnswarden", "https://doh.eu.dnswarden.com/adblock", "Germany/Adblock"),
    ("dnswarden", "https://doh.eu.dnswarden.com/uncensored", "Germany/Uncensored"),
    ("dnswarden", "https://doh.eu.dnswarden.com/adultfilter", "Germany/AdultFilter"),
    ("dnswarden", "https://doh.us.dnswarden.com/adblock", "USA/Adblock"),
    ("dnswarden", "https://doh.us.dnswarden.com/uncensored", "USA/Uncensored"),
    ("dnswarden", "https://doh.us.dnswarden.com/adultfilter", "USA/AdultFilter"),
    ("EdgyDNS", "https://dns.edgy.network/dns-query", "-"),
    ("e-utp.net", "https://dnscache.e-utp.net/dns-query", "-"),
    ("ffmuc.net", "https://doh.ffmuc.net/dns-query", "-"),
    ("Foundation for Applied Privacy", "https://doh.applied-privacy.net/query", "-"),
    ("Google", "https://dns.google/dns-query", "Main"),
    ("Google", "https://dns64.dns.google/dns-query", "DNS64"),
    ("Hostux.net", "https://dns.hostux.net/dns-query", "Uncensored"),
    ("Hostux.net", "https://dns.hostux.net/ads", "Adblocking"),
    ("Hurricane Electric", "https://ordns.he.net/dns-query", "-"),
    ("jitender", "https://jit.ddns.net/dns-query", "-"),
    ("jp.tiar.app", "https://jp.tiar.app/dns-query", "-"),
    ("jp.tiarap.org", "https://jp.tiarap.org/dns-query", "-"),
    ("LavaDNS", "https://us1.dns.lavate.ch/dns-query", "USA"),
    ("LavaDNS", "https://eu1.dns.lavate.ch/dns-query", "Finland"),
    ("LibreDNS", "https://doh.libredns.gr/dns-query", "-"),
    ("Mullvad", "https://doh.mullvad.net/dns-query", "Non-blocking"),
    ("Mullvad", "https://adblock.doh.mullvad.net/dns-query", "Adblocking"),
    ("Moulticast", "https://dns.moulticast.net/dns-query", "-"),
    ("NekomimiRouter", "https://dns.dns-over-https.com/dns-query", "-"),
    ("pf-doh", "https://doh.post-factum.tk/dns-query", "-"),
    ("Plan9-dns", "https://hydra.plan9-ns1.com/dns-query", "New Jersey"),
    ("Plan9-dns", "https://draco.plan9-ns2.com/dns-query", "Florida"),
    ("PowerDNS", "https://doh.powerdns.org", "-"),
    ("Quad9", "https://dns.quad9.net/dns-query", "Recommended"),
    ("Quad9", "https://dns9.quad9.net/dns-query", "Secured"),
    ("Quad9", "https://dns10.quad9.net/dns-query", "Unsecured"),
    ("Quad9", "https://dns11.quad9.net/dns-query", "Secured w/ECS"),
    ("RethinkDNS", "https://basic.rethinkdns.com/dns-query", "Non-filtering"),
    ("Rubyfish.cn", "https://dns.rubyfish.cn/dns-query", "-"),
    ("Snopyta", "https://fi.doh.dns.snopyta.org/dns-query", "-"),
    ("SWITCH", "https://dns.switch.ch/dns-query", "-"),
    ("Tiarap", "https://doh.tiar.app/dns-query", "-"),
    ("Tiarap", "https://doh.tiarap.org/dns-query", "-"),
    ("This.web.id", "https://doh.this.web.id/dns-query", "-"),
    ("TWNIC", "https://dns.twnic.tw/dns-query", "-"),
    ("Usable Privacy", "https://adfree.usableprivacy.net/dns-query", "-"),
    ("WeDNS", "https://dns.wevpn.com/dns-query", "Non-blocking"),
    ("WeDNS", "https://dns-weblock.wevpn.com/dns-query", "Ad/malware block"),
    ("@jedisct1", "https://doh.crypto.sx/dns-query", "-"),
    ("mydns.network", "https://freedom.mydns.network/dns-query", "No blocking"),
    ("mydns.network", "https://adblock.mydns.network/dns-query", "Adblock"),
    ("@null31", "https://ibuki.cgnat.net/dns-query", "-"),
    ("dns.seby.io", "https://doh-2.seby.io/dns-query", "-"),
    ("dns.seby.io", "https://doh.seby.io:8443/dns-query", "-"),
]


# ──────────────────────────────────────────────────────────────────────────────
# RFC 1035 WIRE-FORMAT DNS ENCODE / DECODE
# ──────────────────────────────────────────────────────────────────────────────

def build_dns_query(domain: str, record_type: int = 1) -> bytes:
    """Build a raw RFC 1035 wire-format DNS query (default: A record)."""
    transaction_id = random.randint(0, 65535)
    header = struct.pack(
        ">HHHHHH",
        transaction_id,
        0x0100,  # standard query, recursion desired
        1,       # QDCOUNT
        0,       # ANCOUNT
        0,       # NSCOUNT
        0,       # ARCOUNT
    )
    question = b""
    for label in domain.strip(".").split("."):
        encoded = label.encode("ascii")
        question += struct.pack("B", len(encoded)) + encoded
    question += b"\x00"                       # root label
    question += struct.pack(">HH", record_type, 1)  # QTYPE, QCLASS=IN
    return header + question


def parse_dns_response(data: bytes) -> List[str]:
    """Parse a raw RFC 1035 wire-format DNS response and extract A-record IPs."""
    if len(data) < 12:
        return []

    qdcount, ancount = struct.unpack(">HH", data[4:8])
    offset = 12

    def skip_name(off: int) -> int:
        while True:
            length = data[off]
            if length == 0:
                return off + 1
            if length & 0xC0 == 0xC0:  # compression pointer
                return off + 2
            off += 1 + length

    for _ in range(qdcount):
        offset = skip_name(offset)
        offset += 4  # QTYPE + QCLASS

    ips: List[str] = []
    for _ in range(ancount):
        offset = skip_name(offset)
        if offset + 10 > len(data):
            break
        rtype, _rclass, _ttl, rdlength = struct.unpack(">HHIH", data[offset:offset + 10])
        offset += 10
        rdata = data[offset:offset + rdlength]
        if rtype == 1 and rdlength == 4:  # A record
            ips.append(".".join(str(b) for b in rdata))
        offset += rdlength

    return ips


# ──────────────────────────────────────────────────────────────────────────────
# DATA MODELS
# ──────────────────────────────────────────────────────────────────────────────

@dataclass
class DomainResult:
    ips: List[str] = field(default_factory=list)
    error: Optional[str] = None


@dataclass
class ServerResult:
    name: str
    url: str
    region: str
    reachable: bool = False
    latency_ms: Optional[float] = None
    http_version: Optional[str] = None
    error: Optional[str] = None
    domains: Dict[str, DomainResult] = field(default_factory=dict)


# ──────────────────────────────────────────────────────────────────────────────
# NETWORKING
# ──────────────────────────────────────────────────────────────────────────────

async def send_doh_query(
    client: httpx.AsyncClient, url: str, domain: str
) -> Tuple[List[str], float, Optional[str], Optional[str]]:
    """Send one wire-format DoH query. Returns (ips, latency_ms, http_version, error)."""
    query = build_dns_query(domain)
    start = time.perf_counter()
    try:
        response = await client.post(
            url,
            content=query,
            headers={
                "Content-Type": "application/dns-message",
                "Accept": "application/dns-message",
            },
            timeout=TIMEOUT_SECONDS,
        )
        latency_ms = (time.perf_counter() - start) * 1000
        http_version = response.http_version
        if response.status_code != 200:
            return [], latency_ms, http_version, f"HTTP {response.status_code}"
        ips = parse_dns_response(response.content)
        return ips, latency_ms, http_version, None
    except Exception as exc:
        latency_ms = (time.perf_counter() - start) * 1000
        return [], latency_ms, None, f"{type(exc).__name__}: {exc}"


async def test_server(
    client: httpx.AsyncClient,
    name: str,
    url: str,
    region: str,
    domains: List[str],
    semaphore: asyncio.Semaphore,
    progress: Progress,
    task_id,
) -> ServerResult:
    result = ServerResult(name=name, url=url, region=region)

    async with semaphore:
        ips, latency_ms, http_version, error = await send_doh_query(client, url, PROBE_DOMAIN)
        result.latency_ms = round(latency_ms, 1) if latency_ms else None
        result.http_version = http_version

        if error:
            result.reachable = False
            result.error = error
            for domain in domains:
                result.domains[domain] = DomainResult(ips=[], error="server unreachable")
        else:
            result.reachable = True
            for domain in domains:
                d_ips, _, _, d_err = await send_doh_query(client, url, domain)
                result.domains[domain] = DomainResult(ips=d_ips, error=d_err)

    progress.update(task_id, advance=1)
    return result


async def run_all_tests(
    servers: List[Tuple[str, str, str]], domains: List[str]
) -> List[ServerResult]:
    semaphore = asyncio.Semaphore(CONCURRENCY_LIMIT)
    console = Console()
    limits = httpx.Limits(
        max_connections=CONCURRENCY_LIMIT, max_keepalive_connections=CONCURRENCY_LIMIT
    )

    async with httpx.AsyncClient(http2=True, limits=limits, follow_redirects=True) as client:
        with Progress(
            SpinnerColumn(),
            TextColumn("[bold cyan]Testing DoH servers[/]"),
            BarColumn(),
            TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
            TextColumn("({task.completed}/{task.total})"),
            TimeElapsedColumn(),
            console=console,
        ) as progress:
            task_id = progress.add_task("test", total=len(servers))
            tasks = [
                test_server(client, name, url, region, domains, semaphore, progress, task_id)
                for name, url, region in servers
            ]
            results = await asyncio.gather(*tasks)

    return list(results)


# ──────────────────────────────────────────────────────────────────────────────
# REPORTING
# ──────────────────────────────────────────────────────────────────────────────

def build_summary_table(results: List[ServerResult]) -> Table:
    table = Table(title="DoH Server Reachability Summary", box=box.ROUNDED)
    table.add_column("Status", justify="center", width=10)
    table.add_column("Provider", style="bold")
    table.add_column("Region")
    table.add_column("URL", overflow="fold")
    table.add_column("Latency", justify="right")
    table.add_column("HTTP", justify="center")

    sorted_results = sorted(results, key=lambda r: (not r.reachable, r.latency_ms or 999999))

    for r in sorted_results:
        if r.reachable:
            status = "[bold green]● OPEN[/]"
            latency = f"{r.latency_ms:.0f} ms" if r.latency_ms else "-"
        else:
            status = "[bold red]✕ BLOCKED[/]"
            latency = "-"
        table.add_row(status, r.name, r.region, r.url, latency, r.http_version or "-")

    return table


def build_domain_detail_table(results: List[ServerResult], domain: str) -> Table:
    table = Table(title=f"Resolution results for [bold]{domain}[/]", box=box.SIMPLE_HEAVY)
    table.add_column("Provider", style="bold")
    table.add_column("Region")
    table.add_column("Resolved IPs")
    table.add_column("Note")

    for r in sorted(results, key=lambda r: r.name):
        dr = r.domains.get(domain)
        if dr is None:
            continue
        if dr.ips:
            ip_str = ", ".join(dr.ips)
            note = ""
        else:
            ip_str = "-"
            note = dr.error or "no A record"
        table.add_row(r.name, r.region, ip_str, note)

    return table


def build_final_working_table(results: List[ServerResult], domains: List[str]) -> Table:
    """
    Final table: ONLY working (OPEN) DoH servers, with resolved IPs
    for every target domain, shown side by side, one column per domain.
    """
    table = Table(
        title="[bold green]✔ WORKING DoH SERVERS — Resolved IPs[/]",
        box=box.HEAVY_HEAD,
        show_lines=True,
    )
    table.add_column("Provider", style="bold cyan")
    table.add_column("DoH URL", overflow="fold")
    for domain in domains:
        table.add_column(domain, overflow="fold")

    working = [r for r in results if r.reachable]
    working = sorted(working, key=lambda r: r.latency_ms or 999999)

    for r in working:
        row = [r.name, r.url]
        for domain in domains:
            dr = r.domains.get(domain)
            if dr and dr.ips:
                row.append(", ".join(dr.ips))
            else:
                row.append("[dim]-[/]")
        table.add_row(*row)

    return table


def print_full_report(results: List[ServerResult], domains: List[str], console: Console) -> None:
    console.print()
    console.print(Rule("[bold]ONCIC DoH Endpoint Test Report[/]", style="cyan"))
    console.print()

    total = len(results)
    open_count = sum(1 for r in results if r.reachable)
    blocked_count = total - open_count

    console.print(
        f"  [bold]{total}[/] servers tested   "
        f"[bold green]{open_count} open[/]   "
        f"[bold red]{blocked_count} blocked[/]"
    )
    console.print()
    console.print(build_summary_table(results))

    for dom in domains:
        console.print()
        console.print(build_domain_detail_table(results, dom))

    console.print()
    console.print(Rule(style="cyan"))
    console.print(
        "  [dim]Legend:[/] [bold green]● OPEN[/] = server reachable   "
        "[bold red]✕ BLOCKED[/] = connection failed or filtered"
    )
    console.print()

    # ── Final section: only working servers + resolved IPs ────────────────
    console.print()
    console.print(Rule("[bold green]FINAL RESULT — Working DoH Servers Only[/]", style="green"))
    console.print()
    console.print(build_final_working_table(results, domains))
    console.print()
    console.print(f"  [bold green]{open_count}[/] out of [bold]{total}[/] DoH servers are working.")
    console.print()


# ──────────────────────────────────────────────────────────────────────────────
# ENTRY POINT
# ──────────────────────────────────────────────────────────────────────────────

def main() -> None:
    parser = argparse.ArgumentParser(
        description="ONCIC - Test DoH (DNS over HTTPS) endpoints and resolve popular domains."
    )
    parser.add_argument(
        "--json",
        metavar="FILE",
        help="Also write full results to a JSON file.",
    )
    parser.add_argument(
        "--domains",
        nargs="+",
        default=DOMAINS_TO_RESOLVE,
        help=(
            "Override the list of domains to resolve "
            "(default: youtube.com whatsapp.com x.com instagram.com telegram.org facebook.com)."
        ),
    )
    args = parser.parse_args()

    console = Console()
    console.print()
    console.print("[bold cyan]ONCIC Network Monitoring - DoH Server Tester[/]")
    console.print("[dim]Created by Medo | Part of ONCIC Platform[/]")
    console.print()

    results = asyncio.run(run_all_tests(DOH_SERVERS, args.domains))
    print_full_report(results, args.domains, console)

    if args.json:
        payload = []
        for r in results:
            payload.append(
                {
                    "name": r.name,
                    "url": r.url,
                    "region": r.region,
                    "reachable": r.reachable,
                    "latency_ms": r.latency_ms,
                    "http_version": r.http_version,
                    "error": r.error,
                    "domains": {
                        d: {"ips": dr.ips, "error": dr.error} for d, dr in r.domains.items()
                    },
                }
            )
        with open(args.json, "w", encoding="utf-8") as f:
            json.dump(payload, f, indent=2, ensure_ascii=False)
        console.print(f"\n  [dim]Results written to[/] [bold]{args.json}[/]")


if __name__ == "__main__":
    main()
