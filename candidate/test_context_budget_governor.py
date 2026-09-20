import unittest
from context_budget_governor import Action, ContextPolicy, ContextState, decide, trim_tool_result


class ContextGovernorTests(unittest.TestCase):
    def test_keep_below_soft_limit(self):
        self.assertEqual(decide(ContextState(1000, 10000)), Action.KEEP)

    def test_soft_trim_at_soft_limit(self):
        self.assertEqual(decide(ContextState(2000, 10000)), Action.SOFT_TRIM)

    def test_compact_at_compact_limit(self):
        self.assertEqual(decide(ContextState(4000, 10000)), Action.FLUSH_AND_COMPACT)

    def test_hard_stop_preserves_reserve(self):
        self.assertEqual(decide(ContextState(8001, 10000)), Action.HARD_STOP)

    def test_hard_stop_at_hard_ratio(self):
        p = ContextPolicy(reserve_tokens=0)
        self.assertEqual(decide(ContextState(9000, 10000), p), Action.HARD_STOP)

    def test_tool_trim_is_bounded_and_keeps_edges(self):
        text = "A" * 1500 + "B" * 1500
        out = trim_tool_result(text)
        self.assertLessEqual(len(out), 2000)
        self.assertTrue(out.startswith("A"))
        self.assertTrue(out.endswith("B"))

    def test_invalid_policy_fails_closed(self):
        with self.assertRaises(ValueError):
            decide(ContextState(1, 10), ContextPolicy(soft_ratio=.5, compact_ratio=.4))

    def test_invalid_state_fails_closed(self):
        with self.assertRaises(ValueError):
            decide(ContextState(-1, 10000))


if __name__ == "__main__":
    unittest.main()
