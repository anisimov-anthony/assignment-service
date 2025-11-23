#!/usr/bin/env python3
import json
import math
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path

INPUT_DIR = Path("k6-results/2025-11-23")
OUTPUT_DIR = Path("results")
OUTPUT_DIR.mkdir(exist_ok=True)


def format_time(ms: float) -> str:
    if ms < 0.001:
        return "0s"
    elif ms < 1:
        return f"{ms * 1000:.2f}µs"
    elif ms < 1000:
        return f"{ms:.2f}ms"
    else:
        return f"{ms / 1000:.2f}s"


def percentile(values, p):
    if not values:
        return 0.0
    values = sorted(values)
    k = (len(values) - 1) * (p / 100.0)
    f = math.floor(k)
    c = math.ceil(k)
    if f == c:
        return values[int(k)]
    d0 = values[int(f)] * (c - k)
    d1 = values[int(c)] * (k - f)
    return d0 + d1


class K6Summary:
    def __init__(self):
        self.real_duration = []
        self.real_failed = 0
        self.real_total = 0

        self.trends = defaultdict(list)
        self.counters = defaultdict(float)
        self.checks = Counter()

    def process_point(self, obj):
        if obj.get("type") != "Point":
            return

        metric = obj["metric"]
        data = obj["data"]
        value = data.get("value", 0)
        tags = data.get("tags", {})
        is_real = tags.get("prep") == "false"

        if metric == "http_reqs" and is_real:
            self.real_total += value
        if metric == "http_req_failed" and is_real:
            self.real_failed += value
        if metric == "http_req_duration" and is_real:
            self.real_duration.append(value)

        if metric in [
            "http_req_blocked",
            "http_req_connecting",
            "http_req_receiving",
            "http_req_sending",
            "http_req_waiting",
            "http_req_tls_handshaking",
            "iteration_duration",
        ]:
            self.trends[metric].append(value)

        if metric == "iterations":
            self.counters["iterations"] += value
        if metric == "vus":
            self.counters["vus"] += value
        if metric == "vus_max":
            self.counters["vus_max"] = max(self.counters.get("vus_max", 0), value)
        if metric == "data_received":
            self.counters["data_received"] += value
        if metric == "data_sent":
            self.counters["data_sent"] += value

        if metric.startswith("checks::"):
            check_name = metric[len("checks::") :]
            passed = value > 0
            self.checks[(check_name, passed)] += 1

    def calc_stats(self, values):
        if not values:
            return {"avg": 0, "min": 0, "med": 0, "max": 0, "p90": 0, "p95": 0}
        return {
            "avg": sum(values) / len(values),
            "min": min(values),
            "med": percentile(values, 50),
            "max": max(values),
            "p90": percentile(values, 90),
            "p95": percentile(values, 95),
        }

    def format_bytes(self, b):
        if b < 1024:
            return f"{b} B"
        elif b < 1024 * 1024:
            return f"{b / 1024:.0f} kB"
        else:
            return f"{b / (1024 * 1024):.1f} MB"

    def generate_summary(self, test_name, duration_sec=None):
        lines = []
        lines.append(f"   {test_name}\n")

        total_checks = sum(self.checks.values())
        passed_checks = sum(v for (n, p), v in self.checks.items() if p)
        check_rate = (passed_checks / total_checks * 100) if total_checks else 100
        lines.append(
            f"     checks.........................: {check_rate:6.2f}% {passed_checks} out of {total_checks}"
        )

        recv = self.counters.get("data_received", 0)
        sent = self.counters.get("data_sent", 0)
        recv_rate = recv / duration_sec if duration_sec else 0
        sent_rate = sent / duration_sec if duration_sec else 0
        lines.append(
            f"     data_received..................: {self.format_bytes(recv)} {self.format_bytes(recv_rate)}/s"
        )
        lines.append(
            f"     data_sent......................: {self.format_bytes(sent)} {self.format_bytes(sent_rate)}/s"
        )

        for metric in [
            "http_req_blocked",
            "http_req_connecting",
            "http_req_receiving",
            "http_req_sending",
            "http_req_waiting",
            "http_req_tls_handshaking",
        ]:
            if metric in self.trends and self.trends[metric]:
                st = self.calc_stats(self.trends[metric])
                lines.append(
                    f"     {metric.ljust(30)}: avg={format_time(st['avg'])} min={format_time(st['min'])} med={format_time(st['med'])} max={format_time(st['max'])} p(90)={format_time(st['p90'])} p(95)={format_time(st['p95'])}"
                )

        if self.real_duration:
            st = self.calc_stats(self.real_duration)
            lines.append(
                f"     http_req_duration..............: avg={format_time(st['avg'])} min={format_time(st['min'])} med={format_time(st['med'])} max={format_time(st['max'])} p(90)={format_time(st['p90'])} p(95)={format_time(st['p95'])}"
            )
            lines.append(
                f"       {{ prep:false }}...............: avg={format_time(st['avg'])} min={format_time(st['min'])} med={format_time(st['med'])} max={format_time(st['max'])} p(90)={format_time(st['p90'])} p(95)={format_time(st['p95'])}"
            )

        total_reqs = self.real_total or 1
        failed_pct = self.real_failed / total_reqs * 100
        lines.append(
            f"     http_req_failed................: {failed_pct:6.2f}% {int(self.real_failed)} out of {int(total_reqs)}"
        )

        req_rate = self.real_total / duration_sec if duration_sec else 0
        iter_rate = (
            self.counters.get("iterations", 0) / duration_sec if duration_sec else 0
        )
        lines.append(
            f"     http_reqs......................: {int(self.real_total):4d} {req_rate:.2f}/s"
        )
        lines.append(
            f"     iterations.....................: {int(self.counters.get('iterations', 0)):4d} {iter_rate:.2f}/s"
        )

        lines.append(f"     vus............................: 0 min=0 max=0")
        lines.append(
            f"     vus_max........................: {int(self.counters.get('vus_max', 1))} min=1 max=1"
        )

        return "\n".join(lines)


def extract_duration(points):
    times = [
        p["data"].get("time")
        for p in points
        if p.get("type") == "Point" and p["data"].get("time")
    ]
    if not times:
        return 60
    try:
        start = datetime.fromisoformat(times[0].replace("Z", "+00:00"))
        end = datetime.fromisoformat(times[-1].replace("Z", "+00:00"))
        return max((end - start).total_seconds(), 1)
    except:
        return 60


def main():
    for json_path in sorted(INPUT_DIR.glob("*_results.json")):
        print(f"Обрабатываю: {json_path.name}")

        summary = K6Summary()
        points = []

        with open(json_path, encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    obj = json.loads(line)
                    points.append(obj)
                    if obj.get("type") == "Point":
                        summary.process_point(obj)
                except json.JSONDecodeError:
                    continue

        duration = extract_duration(points)

        name = json_path.stem.replace("_results", "")
        pretty = name.replace("1k_", "").replace("sli_", "").replace("_", " ").title()
        if name.startswith("1k_"):
            pretty += " (1k rps)"
        elif name.startswith("sli_"):
            pretty += " (SLI)"

        text = summary.generate_summary(pretty, duration)
        out_file = OUTPUT_DIR / f"{name}.txt"
        out_file.write_text(text + "\n", encoding="utf-8")
        print(f"   Готово → {out_file.name}\n")


if __name__ == "__main__":
    main()
