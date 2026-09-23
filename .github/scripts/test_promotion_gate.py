import unittest

import promotion_gate as pg


class PromotionGateTests(unittest.TestCase):
    def test_candidate_to_main_allowed(self):
        self.assertEqual(
            pg.validate_transition("candidate/x", "main"),
            "candidate_to_main",
        )

    def test_non_candidate_to_main_blocked(self):
        with self.assertRaises(pg.GateError):
            pg.validate_transition("feature/x", "main")

    def test_main_to_stable_allowed(self):
        self.assertEqual(
            pg.validate_transition("main", "stable"),
            "main_to_stable",
        )

    def test_candidate_to_stable_blocked(self):
        with self.assertRaises(pg.GateError):
            pg.validate_transition("candidate/x", "stable")

    def test_linux_and_windows_required(self):
        self.assertTrue(pg.jobs_cover_linux_windows([
            {"name": "test (ubuntu-latest)", "conclusion": "success"},
            {"name": "test (windows-latest)", "conclusion": "success"},
        ]))
        self.assertFalse(pg.jobs_cover_linux_windows([
            {"name": "test (ubuntu-latest)", "conclusion": "success"},
        ]))

    def test_rollback_must_point_to_exact_target(self):
        refs = [
            {"ref": "refs/heads/rollback/a", "object": {"sha": "old"}},
            {"ref": "refs/heads/rollback/b", "object": {"sha": "target"}},
        ]
        self.assertEqual(pg.matching_rollback(refs, "target"), "rollback/b")
        self.assertIsNone(pg.matching_rollback(refs, "missing"))

    def test_candidate_evidence_accepts_validation(self):
        runs = [{
            "id": 10,
            "name": "NEURA Verified Action Loop Validation",
            "event": "push",
            "conclusion": "success",
            "head_sha": "abc",
        }]
        ev = pg.select_evidence(
            runs,
            "main",
            "abc",
            lambda _: [
                {"name": "test (ubuntu-latest)", "conclusion": "success"},
                {"name": "test (windows-latest)", "conclusion": "success"},
            ],
        )
        self.assertEqual(ev["run_id"], 10)

    def test_stable_requires_promotion_gate_push(self):
        runs = [{
            "id": 11,
            "name": "NEURA Promotion Gate",
            "event": "pull_request",
            "conclusion": "success",
            "head_sha": "abc",
        }]
        ev = pg.select_evidence(
            runs,
            "stable",
            "abc",
            lambda _: [
                {"name": "test (ubuntu-latest)", "conclusion": "success"},
                {"name": "test (windows-latest)", "conclusion": "success"},
            ],
        )
        self.assertIsNone(ev)

    def test_failed_job_invalidates_evidence(self):
        runs = [{
            "id": 12,
            "name": "NEURA Promotion Gate",
            "event": "push",
            "conclusion": "success",
            "head_sha": "abc",
        }]
        ev = pg.select_evidence(
            runs,
            "stable",
            "abc",
            lambda _: [
                {"name": "test (ubuntu-latest)", "conclusion": "success"},
                {"name": "test (windows-latest)", "conclusion": "failure"},
            ],
        )
        self.assertIsNone(ev)


if __name__ == "__main__":
    unittest.main()
