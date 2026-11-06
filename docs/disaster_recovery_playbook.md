# Disaster Recovery Playbook

## Purpose

A backup strategy that does not test restores regularly is a vulnerability. Database dumps fail for multiple reasons: truncated uploads, storage bit-rot, unsupported major version upgrades, or missing extensions. This playbook defines how to use `restorecraft` to establish measurable recovery confidence.

## Core Metrics

### Recovery Time Objective (RTO)
RTO is the target duration within which a system must be restored after a failure. `restorecraft` measures RTO on every drill run by recording the exact millisecond timestamps from initial download through final constraint verification:

$$\text{Total RTO} = T_{\text{fetch}} + T_{\text{sandbox\_boot}} + T_{\text{restore}} + T_{\text{audit}}$$

If the measured RTO exceeds the configured SLA, `restorecraft` flags the drill with an SLA overrun penalty and sends an alert.

### Recovery Health Grade
Drills are graded from A+ down to F based on the following formula:

$$\text{Health Score} = 100 - (\text{DataLossPenalty} \times 50) - (\text{ConstraintViolations} \times 20) - (\text{RTO Overrun Penalty})$$

- **Grade A+ (95 to 100)**: Clean restore, zero constraint failures, RTO comfortably below target SLA.
- **Grade A (85 to 94)**: Clean restore, zero constraint failures, RTO within SLA.
- **Grade B (70 to 84)**: Restored successfully, but warnings detected or RTO was within 25% of SLA threshold.
- **Grade C (50 to 69)**: Restored successfully, but foreign key or parity violations were identified.
- **Grade F (0 to 49)**: Restoration process aborted, file corrupted, or total execution timed out.

## Establishing a Production Routine

1. **Daily Off-Peak Execution**:
   Configure `restorecraft` to pull the latest backup daily during low-traffic windows (e.g. 03:00 UTC).

2. **Automated SOC 2 and ISO 27001 Audit Evidence**:
   For each completed drill, `restorecraft` issues a cryptographically signed JSON certificate. Archive these certificates in a compliance repository or verify them via:
   ```bash
   restorecraft verify-cert --cert ./certificates/cert-drill-123.json --secret $RESTORECRAFT_SIGNING_SECRET
   ```

3. **Chaos Testing with Corrupted Fixtures**:
   Regularly test your alerting pipeline by pointing `restorecraft` at an intentionally damaged backup. Use the built-in generator:
   ```bash
   restorecraft fixtures generate --corrupt
   ```
   Confirm that your monitoring channels receive an immediate Grade F alert.
