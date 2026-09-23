package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (c *Core) ExecutePlan(ctx context.Context, goal string, plan Plan) CommandResult {
	res := CommandResult{Goal: strings.TrimSpace(goal), Plan: &plan}
	if c == nil || c.tools == nil {
		res.Status = "error"
		res.Message = "core tools unavailable"
		return res
	}
	checked, err := validatePlan(plan, c.tools.List())
	if err != nil {
		res.Status = "blocked"
		res.Message = "plan rejected before execution: " + err.Error()
		if c.recorder != nil {
			_, _ = c.recorder.Record("", goal, "plan", "plan", "blocked", err)
		}
		return res
	}
	if len(checked.Steps) == 0 {
		res.Status = "ok"
		res.Message = "plan contains no actions"
		if c.recorder != nil {
			_, _ = c.recorder.Record("", goal, "plan", "plan", "verified", nil)
		}
		return res
	}
	if c.recorder != nil {
		_, _ = c.recorder.Record("", goal, "plan", "plan", "verified", nil)
	}

	res.Status = "ok"
	for i, step := range checked.Steps {
		if err := ctx.Err(); err != nil {
			res.Status = "partial"
			res.Message = "execution cancelled"
			if c.recorder != nil {
				_, _ = c.recorder.Record("", goal, step.Tool, "execute", "cancelled", err)
			}
			break
		}
		started := time.Now().UTC()
		if c.recorder != nil {
			_, _ = c.recorder.Record("", goal, step.Tool, "authorize", "requested", nil)
		}
		tr := c.runTool(ctx, goal, "verified-plan", step.Tool, step.Input)
		res.Results = append(res.Results, tr)
		if tr.Error != "" {
			res.Status = "partial"
			res.Message = fmt.Sprintf("step %d blocked: %s", i+1, tr.Error)
			res.ReceiptIDs = append(res.ReceiptIDs, c.receipt(goal, step.Tool, "blocked", "", tr.Error, started))
			if c.recorder != nil {
				_, _ = c.recorder.Record("", goal, step.Tool, "execute", "blocked", errors.New(tr.Error))
			}
			break
		}
		if !tr.Verified {
			recovery := ""
			if step.Tool == "fs.write.workspace" && c.tools.writer != nil {
				if wr, ok := tr.Data.(WriteReceipt); ok {
					if rbErr := c.tools.writer.Rollback(wr); rbErr != nil {
						recovery = "; rollback failed: " + rbErr.Error()
					} else {
						recovery = "; rollback verified"
					}
				}
			}
			errText := "postcondition not verified" + recovery
			res.Status = "partial"
			res.Message = fmt.Sprintf("step %d verification failed", i+1)
			res.ReceiptIDs = append(res.ReceiptIDs, c.receipt(goal, step.Tool, "failed", "", errText, started))
			if c.recorder != nil {
				_, _ = c.recorder.Record("", goal, step.Tool, "verify", "failed", errors.New(errText))
			}
			break
		}
		res.ReceiptIDs = append(res.ReceiptIDs, c.receipt(goal, step.Tool, "verified", "postcondition verified", "", started))
		if c.recorder != nil {
			_, _ = c.recorder.Record("", goal, step.Tool, "verify", "verified", nil)
		}
	}
	if res.Status == "ok" {
		res.Message = "plan executed with verified effects"
	}
	return res
}
