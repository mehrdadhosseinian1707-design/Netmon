#!/usr/bin/env python3
"""
ONCIC Network Monitoring - DNS Server Reachability Tester
Tests reachability and response time of public DNS servers worldwide

Tests 20 major DNS providers with their primary and secondary servers.
Part of ONCIC Network Monitoring Platform / Created by Medo
"""

import argparse
import asyncio
import json
import socket
import struct
import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple

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

# Test domains for DNS resolution
TEST_DOMAINS = [
    "google.com",
    "cloudflare.com",
    "youtube.com",
]

TIMEOUT_SECONDS = 5.0
DNS_PORT = 53

# DNS Servers Database
DNS_SERVERS = [
    (1, "Google Public DNS", "8.8.8.8", "8.8.4.4"),
    (2, "Cloudflare DNS", "1.1.1.1", "1.0.0.1"),
    (3, "Quad9", "9.9.9.9", "149.112.112.112"),
    (4, "OpenDNS (Cisco)", "208.67.222.222", "208.67.220.220"),
    (5, "AdGuard DNS", "94.140.14.14", "94.140.15.15"),
    (6, "CleanBrowsing", "185.228.168.168", "185.228.169.168"),
    (7, "Comodo Secure DNS", "8.26.56.26", "8.20.247.20"),
    (8, "Quad101", "101.101.101.101", "101.102.103.104"),
    (9, "OpenNIC", "217.160.70.42", None),
    (10, "SWITCH DNS", "130.59.31.248", None),
    (11, "DNS.SB", "185.222.222.222", "45.11.45.11"),
    (12, "Verisign Public DNS", "64.6.64.6", "64.6.65.6"),
    (13, "Level3", "4.2.2.1", "4.2.2.2"),
    (14, "Neustar UltraRecursive", "64.6.64.6", "64.6.65.6"),
    (15, "Control D", "76.76.2.0", "76.76.10.0"),
    (16, "Alternate DNS", "76.76.19.19", "76.223.122.150"),
    (17, "Yandex DNS", "77.88.8.8", "77.88.8.1"),
    (18, "AliDNS", "223.5.5.5", "223.6.6.6"),
    (19, "DNSPod", "119.29.29.29", "182.254.116.116"),
    (20, "114DNS", "114.114.114.114", "114.114.115.115"),
]


# ──────────────────────────────────────────────────────────────────────────────
# DNS PROTOCOL (RFC 1035)
# ──────────────────────────────────────────────────────────────────────────────

def build_dns_query(domain: str, query_type: int = 1) -> bytes:
    """Build a DNS query packet (A record by default)."""
    import random

    transaction_id = random.randint(0, 65535)
    flags = 0x0100  # Standard query with recursion

    # Header: ID, Flags, QDCOUNT, ANCOUNT, NSCOUNT, ARCOUNT
    header = struct.pack(">HHHHHH", transaction_id, flags, 1, 0, 0, 0)

    # Question section
    question = b""
    for label in domain.strip(".").split("."):
        encoded = label.encode("ascii")
        question += struct.pack("B", len(encoded)) + encoded
    question += b"\x00"  # End of name
    question += struct.pack(">HH", query_type, 1)  # Type A, Class IN

    return header + question


def parse_dns_response(data: bytes) -> Tuple[bool, List[str]]:
    """
    Parse DNS response packet.
    Returns (success, list_of_ips)
    """
    if len(data) < 12:
        return False, []

    # Parse header
    _transaction_id = struct.unpack(">H", data[0:2])[0]
    flags = struct.unpack(">H", data[2:4])[0]
    qdcount = struct.unpack(">H", data[4:6])[0]
    ancount = struct.unpack(">H", data[6:8])[0]

    # Check if response
    if not (flags & 0x8000):
        return False, []

    # Check response code
    rcode = flags & 0x000F
    if rcode != 0:  # No error
        return False, []

    offset = 12

    # Skip question section
    def skip_name(off: int) -> int:
        while True:
            if off >= len(data):
                return off
            length = data[off]
            if length == 0:
                return off + 1
            if length & 0xC0 == 0xC0:  # Compression pointer
                return off + 2
            off += 1 + length

    for _ in range(qdcount):
        offset = skip_name(offset)
        offset += 4  # Type + Class

    # Parse answer section
    ips = []
    for _ in range(ancount):
        offset = skip_name(offset)

        if offset + 10 > len(data):
            break

        rtype, _rclass, _ttl, rdlength = struct.unpack(">HHIH", data[offset:offset + 10])
        offset += 10

        if rtype == 1 and rdlength == 4:  # A record
            ip = ".".join(str(b) for b in data[offset:offset + 4])
            ips.append(ip)

        offset += rdlength

    return len(ips) > 0, ips


# ──────────────────────────────────────────────────────────────────────────────
# DATA MODELS
# ──────────────────────────────────────────────────────────────────────────────

@dataclass
class DNSTestResult:
    ip: str
    reachable: bool = False
    latency_ms: Optional[float] = None
    resolved_ips: List[str] = field(default_factory=list)
    error: Optional[str] = None


@dataclass
class DNSProviderResult:
    rank: int
    name: str
    primary: DNSTestResult
    secondary: Optional[DNSTestResult] = None


# ──────────────────────────────────────────────────────────────────────────────
# DNS TESTING
# ──────────────────────────────────────────────────────────────────────────────

async def test_dns_server(ip: str, domain: str) -> Tuple[bool, float, List[str], Optional[str]]:
    """
    Test a single DNS server by resolving a domain.
    Returns (success, latency_ms, resolved_ips, error)
    """
    loop = asyncio.get_event_loop()

    try:
        # Create UDP socket
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.settimeout(TIMEOUT_SECONDS)

        # Build and send query
        query = build_dns_query(domain)

        start = time.perf_counter()

        # Send query
        await loop.run_in_executor(None, sock.sendto, query, (ip, DNS_PORT))

        # Receive response
        data, _ = await loop.run_in_executor(None, sock.recvfrom, 512)

        latency = (time.perf_counter() - start) * 1000

        sock.close()

        # Parse response
        success, ips = parse_dns_response(data)

        if success:
            return True, latency, ips, None
        else:
            return False, latency, [], "Invalid response"

    except socket.timeout:
        return False, TIMEOUT_SECONDS * 1000, [], "Timeout"
    except Exception as e:
        return False, 0, [], f"{type(e).__name__}: {str(e)[:50]}"


async def test_dns_provider(
    rank: int,
    name: str,
    primary: str,
    secondary: Optional[str],
    test_domain: str
) -> DNSProviderResult:
    """Test a DNS provider (primary and secondary servers)."""

    # Test primary
    success, latency, ips, error = await test_dns_server(primary, test_domain)
    primary_result = DNSTestResult(
        ip=primary,
        reachable=success,
        latency_ms=round(latency, 1) if success else None,
        resolved_ips=ips,
        error=error
    )

    # Test secondary if exists
    secondary_result = None
    if secondary:
        success, latency, ips, error = await test_dns_server(secondary, test_domain)
        secondary_result = DNSTestResult(
            ip=secondary,
            reachable=success,
            latency_ms=round(latency, 1) if success else None,
            resolved_ips=ips,
            error=error
        )

    return DNSProviderResult(
        rank=rank,
        name=name,
        primary=primary_result,
        secondary=secondary_result
    )


async def test_all_dns_servers(servers: List[Tuple], test_domain: str) -> List[DNSProviderResult]:
    """Test all DNS servers concurrently."""
    console = Console()

    with Progress(
        SpinnerColumn(),
        TextColumn("[bold cyan]Testing DNS servers[/]"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
        TextColumn("({task.completed}/{task.total})"),
        TimeElapsedColumn(),
        console=console,
    ) as progress:
        task_id = progress.add_task("test", total=len(servers))

        tasks = []
        for rank, name, primary, secondary in servers:
            task = test_dns_provider(rank, name, primary, secondary, test_domain)
            tasks.append(task)

        results = []
        for coro in asyncio.as_completed(tasks):
            result = await coro
            results.append(result)
            progress.update(task_id, advance=1)

    # Sort by rank
    results.sort(key=lambda r: r.rank)
    return results


# ──────────────────────────────────────────────────────────────────────────────
# REPORTING
# ──────────────────────────────────────────────────────────────────────────────

def build_summary_table(results: List[DNSProviderResult]) -> Table:
    """Build summary table with all DNS servers."""
    table = Table(
        title="DNS Server Reachability Test Results",
        box=box.ROUNDED,
        show_lines=True
    )

    table.add_column("#", justify="center", style="dim", width=4)
    table.add_column("Provider", style="bold cyan")
    table.add_column("Primary IP", style="yellow")
    table.add_column("Status", justify="center", width=12)
    table.add_column("Latency", justify="right", width=10)
    table.add_column("Secondary IP", style="yellow")
    table.add_column("Status", justify="center", width=12)
    table.add_column("Latency", justify="right", width=10)

    for r in results:
        # Primary status
        if r.primary.reachable:
            p_status = "[bold green]✓ ONLINE[/]"
            p_latency = f"{r.primary.latency_ms:.0f} ms"
        else:
            p_status = "[bold red]✗ FAILED[/]"
            p_latency = "-"

        # Secondary status
        if r.secondary:
            if r.secondary.reachable:
                s_status = "[bold green]✓ ONLINE[/]"
                s_latency = f"{r.secondary.latency_ms:.0f} ms"
            else:
                s_status = "[bold red]✗ FAILED[/]"
                s_latency = "-"
        else:
            s_status = "[dim]—[/]"
            s_latency = "[dim]—[/]"

        table.add_row(
            str(r.rank),
            r.name,
            r.primary.ip,
            p_status,
            p_latency,
            r.secondary.ip if r.secondary else "[dim]—[/]",
            s_status,
            s_latency
        )

    return table


def build_working_servers_table(results: List[DNSProviderResult]) -> Table:
    """Build table with only working DNS servers, sorted by latency."""
    table = Table(
        title="[bold green]✔ WORKING DNS SERVERS (Sorted by Speed)[/]",
        box=box.HEAVY_HEAD
    )

    table.add_column("Rank", justify="center", style="bold")
    table.add_column("Provider", style="bold cyan")
    table.add_column("Server IP", style="yellow")
    table.add_column("Type", justify="center")
    table.add_column("Latency", justify="right", style="green")
    table.add_column("Resolved IPs", overflow="fold")

    # Collect all working servers
    working = []
    for r in results:
        if r.primary.reachable:
            working.append((
                r.rank,
                r.name,
                r.primary.ip,
                "Primary",
                r.primary.latency_ms,
                ", ".join(r.primary.resolved_ips)
            ))
        if r.secondary and r.secondary.reachable:
            working.append((
                r.rank,
                r.name,
                r.secondary.ip,
                "Secondary",
                r.secondary.latency_ms,
                ", ".join(r.secondary.resolved_ips)
            ))

    # Sort by latency (fastest first)
    working.sort(key=lambda x: x[4] if x[4] else 99999)

    for rank, name, ip, server_type, latency, ips in working:
        table.add_row(
            str(rank),
            name,
            ip,
            server_type,
            f"{latency:.0f} ms" if latency else "-",
            ips if ips else "[dim]—[/]"
        )

    return table


def build_statistics_table(results: List[DNSProviderResult]) -> Table:
    """Build statistics summary table."""
    table = Table(title="Statistics Summary", box=box.SIMPLE)

    total_servers = len(results) * 2 - sum(1 for r in results if r.secondary is None)
    working_primary = sum(1 for r in results if r.primary.reachable)
    working_secondary = sum(1 for r in results if r.secondary and r.secondary.reachable)
    total_working = working_primary + working_secondary

    # Calculate average latency of working servers
    latencies = []
    for r in results:
        if r.primary.reachable and r.primary.latency_ms:
            latencies.append(r.primary.latency_ms)
        if r.secondary and r.secondary.reachable and r.secondary.latency_ms:
            latencies.append(r.secondary.latency_ms)

    avg_latency = sum(latencies) / len(latencies) if latencies else 0
    min_latency = min(latencies) if latencies else 0

    table.add_column("Metric", style="bold")
    table.add_column("Value", justify="right", style="cyan")

    table.add_row("Total DNS Providers", str(len(results)))
    table.add_row("Total Servers Tested", str(total_servers))
    table.add_row("Working Servers", f"[bold green]{total_working}[/]")
    table.add_row("Failed Servers", f"[bold red]{total_servers - total_working}[/]")
    table.add_row("Success Rate", f"{(total_working / total_servers * 100):.1f}%")
    table.add_row("Average Latency", f"{avg_latency:.1f} ms")
    table.add_row("Fastest Server", f"{min_latency:.1f} ms")

    return table


def print_full_report(results: List[DNSProviderResult], test_domain: str, console: Console):
    """Print complete test report."""
    console.print()
    console.print(Rule("[bold]ONCIC DNS Server Reachability Report[/]", style="cyan"))
    console.print()
    console.print(f"  Test Domain: [bold]{test_domain}[/]")
    console.print(f"  Timeout: {TIMEOUT_SECONDS}s")
    console.print()

    # Summary table
    console.print(build_summary_table(results))
    console.print()

    # Statistics
    console.print(build_statistics_table(results))
    console.print()

    # Working servers only (sorted by speed)
    console.print(build_working_servers_table(results))
    console.print()

    console.print(Rule(style="cyan"))
    console.print(
        "  [dim]Legend:[/] [bold green]✓ ONLINE[/] = server responsive   "
        "[bold red]✗ FAILED[/] = timeout or error"
    )
    console.print()


# ──────────────────────────────────────────────────────────────────────────────
# ENTRY POINT
# ──────────────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="ONCIC - Test reachability and speed of public DNS servers"
    )
    parser.add_argument(
        "--domain",
        default="google.com",
        help="Domain to resolve for testing (default: google.com)"
    )
    parser.add_argument(
        "--json",
        metavar="FILE",
        help="Export results to JSON file"
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=TIMEOUT_SECONDS,
        help=f"Timeout in seconds (default: {TIMEOUT_SECONDS})"
    )
    args = parser.parse_args()

    global TIMEOUT_SECONDS
    TIMEOUT_SECONDS = args.timeout

    console = Console()
    console.print()
    console.print("[bold cyan]ONCIC Network Monitoring - DNS Server Reachability Tester[/]")
    console.print("[dim]Created by Medo | Part of ONCIC Platform[/]")
    console.print()

    # Run tests
    results = asyncio.run(test_all_dns_servers(DNS_SERVERS, args.domain))

    # Print report
    print_full_report(results, args.domain, console)

    # Export JSON if requested
    if args.json:
        export_data = []
        for r in results:
            provider_data = {
                "rank": r.rank,
                "name": r.name,
                "primary": {
                    "ip": r.primary.ip,
                    "reachable": r.primary.reachable,
                    "latency_ms": r.primary.latency_ms,
                    "resolved_ips": r.primary.resolved_ips,
                    "error": r.primary.error
                }
            }

            if r.secondary:
                provider_data["secondary"] = {
                    "ip": r.secondary.ip,
                    "reachable": r.secondary.reachable,
                    "latency_ms": r.secondary.latency_ms,
                    "resolved_ips": r.secondary.resolved_ips,
                    "error": r.secondary.error
                }

            export_data.append(provider_data)

        with open(args.json, "w", encoding="utf-8") as f:
            json.dump(export_data, f, indent=2, ensure_ascii=False)

        console.print(f"  [dim]Results exported to[/] [bold]{args.json}[/]")
        console.print()


if __name__ == "__main__":
    main()
