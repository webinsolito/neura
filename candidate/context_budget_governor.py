"""NEURA C9.50B candidate: deterministic context-budget governor.

No model calls, network access, or external dependencies. The governor only
calculates budgets and selects deterministic compaction actions. Persistence
and summarisation are deliberately delegated to separately validated layers.
"""
from dataclasses import dataclass
from enum import Enum


class Action(str, Enum):
    KEEP = "keep"
    SOFT_TRIM = "soft_trim"
    FLUSH_AND_COMPACT = "flush_and_compact"
    HARD_STOP = "hard_stop"


@dataclass(frozen=True)
class ContextPolicy:
    soft_ratio: float = 0.20
    compact_ratio: float = 0.40
    hard_ratio: float = 0.90
    reserve_tokens: int = 2048
    max_tool_chars: int = 2000
    max_bootstrap_chars: int = 8000

    def validate(self) -> None:
        if not (0 < self.soft_ratio < self.compact_ratio < self.hard_ratio <= 1):
            raise ValueError("ratios must be ordered: soft < compact < hard <= 1")
        if self.reserve_tokens < 0 or self.max_tool_chars <= 0 or self.max_bootstrap_chars <= 0:
            raise ValueError("budgets must be positive")


@dataclass(frozen=True)
class ContextState:
    used_tokens: int
    max_tokens: int

    @property
    def ratio(self) -> float:
        if self.max_tokens <= 0 or self.used_tokens < 0:
            raise ValueError("invalid context state")
        return self.used_tokens / self.max_tokens


def decide(state: ContextState, policy: ContextPolicy = ContextPolicy()) -> Action:
    """Return the safest deterministic action for the current context usage."""
    policy.validate()
    ratio = state.ratio
    remaining = state.max_tokens - state.used_tokens
    if remaining < policy.reserve_tokens or ratio >= policy.hard_ratio:
        return Action.HARD_STOP
    if ratio >= policy.compact_ratio:
        return Action.FLUSH_AND_COMPACT
    if ratio >= policy.soft_ratio:
        return Action.SOFT_TRIM
    return Action.KEEP


def trim_tool_result(text: str, policy: ContextPolicy = ContextPolicy()) -> str:
    """Bound tool output while retaining both beginning and end for diagnostics."""
    policy.validate()
    limit = policy.max_tool_chars
    if len(text) <= limit:
        return text
    marker = "\n…[NEURA deterministic trim]…\n"
    usable = max(0, limit - len(marker))
    head = usable // 2
    tail = usable - head
    return text[:head] + marker + text[-tail:] if tail else text[:head] + marker
