#!/usr/bin/env python3
import argparse
import json
import sys
import urllib.parse
import urllib.request


class GateError(RuntimeError):
    pass


def clean_ref(ref):
    ref = (ref or "").strip()
    return ref[len("refs/heads/"):] if ref.startswith("refs/heads/") else ref


def validate_transition(source_ref, target_ref):
    source = clean_ref(source_ref)
    target = clean_ref(target_ref)
    if target == "main":
        if not source.startswith("candidate/"):
            raise GateError("main accepts promotion only from candidate/*")
        return "candidate_to_main"
    if target == "stable":
        if source != "main":
            raise GateError("stable accepts promotion only from main")
        return "main_to_stable"
    raise GateError("unsupported promotion target")


def jobs_cover_linux_windows(jobs):
    linux = False
    windows = False
    for job in jobs:
        if job.get("conclusion") != "success":
            continue
        name = (job.get("name") or "").lower()
        linux = linux or "ubuntu" in name or "linux" in name
        windows = windows or "windows" in name
    return linux and windows


def run_allowed(run, target_ref):
    if run.get("conclusion") != "success":
        return False
    name = run.get("name") or ""
    target = clean_ref(target_ref)
    if target == "stable":
        return name == "NEURA Promotion Gate" and run.get("event") == "push"
    return name == "NEURA Promotion Gate" or name.endswith("Validation")


def matching_rollback(refs, target_sha):
    for ref in refs:
        name = ref.get("ref") or ""
        sha = (ref.get("object") or {}).get("sha")
        if name.startswith("refs/heads/rollback/") and sha == target_sha:
            return name[len("refs/heads/"):]
    return None


def select_evidence(runs, target_ref, source_sha, jobs_loader):
    for run in runs:
        if run.get("head_sha") != source_sha or not run_allowed(run, target_ref):
            continue
        jobs = jobs_loader(run["id"])
        if jobs_cover_linux_windows(jobs):
            return {
                "run_id": run["id"],
                "workflow": run.get("name"),
                "event": run.get("event"),
                "html_url": run.get("html_url"),
            }
    return None


class GitHubAPI:
    def __init__(self, repo, token):
        self.repo = repo
        self.token = token

    def get(self, path):
        req = urllib.request.Request(
            f"https://api.github.com/repos/{self.repo}{path}",
            headers={
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {self.token}",
                "X-GitHub-Api-Version": "2022-11-28",
                "User-Agent": "neura-promotion-gate",
            },
        )
        with urllib.request.urlopen(req, timeout=20) as resp:
            return json.load(resp)


def verify(args):
    transition = validate_transition(args.source_ref, args.target_ref)
    api = GitHubAPI(args.repo, args.token)

    refs = api.get("/git/matching-refs/heads/rollback/")
    rollback = matching_rollback(refs, args.target_sha)
    if not rollback:
        raise GateError("no rollback/* ref points to current target SHA")

    query = urllib.parse.urlencode(
        {"head_sha": args.source_sha, "status": "completed", "per_page": 100}
    )
    runs = api.get(f"/actions/runs?{query}").get("workflow_runs", [])

    def jobs_loader(run_id):
        return api.get(f"/actions/runs/{run_id}/jobs?per_page=100").get("jobs", [])

    evidence = select_evidence(runs, args.target_ref, args.source_sha, jobs_loader)
    if not evidence:
        if clean_ref(args.target_ref) == "stable":
            raise GateError("main SHA lacks a successful push NEURA Promotion Gate with Linux + Windows evidence")
        raise GateError("candidate SHA lacks a successful validation run with Linux + Windows evidence")

    return {
        "ok": True,
        "transition": transition,
        "source_ref": clean_ref(args.source_ref),
        "target_ref": clean_ref(args.target_ref),
        "source_sha": args.source_sha,
        "target_sha": args.target_sha,
        "rollback_ref": rollback,
        "evidence": evidence,
    }


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--repo", required=True)
    p.add_argument("--source-ref", required=True)
    p.add_argument("--target-ref", required=True)
    p.add_argument("--source-sha", required=True)
    p.add_argument("--target-sha", required=True)
    p.add_argument("--token", required=True)
    args = p.parse_args()
    try:
        result = verify(args)
    except Exception as exc:
        print(f"PROMOTION BLOCKED: {exc}", file=sys.stderr)
        return 1
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
