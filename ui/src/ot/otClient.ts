/**
 * OTClient coordinates between the Tiptap editor and the server's OT engine.
 *
 * Lifecycle:
 *  1. Editor produces local steps → call localSteps() to get a WS message payload.
 *  2. Server responds with ot_ack → call handleMessage() to advance the confirmed version.
 *  3. Server broadcasts ot_steps from another client → call handleMessage() to deliver
 *     remote steps to the editor via the onSteps callback.
 *  4. Server sends ot_rebase (version conflict) → call handleMessage() to re-apply buffered
 *     local steps on top of the server's version.
 */
export class OTClient {
  private readonly docId: string
  private _version: number
  private inflight: unknown[]
  private readonly onSteps: (steps: unknown[]) => void

  constructor(
    docId: string,
    initialVersion: number,
    onSteps: (steps: unknown[]) => void,
  ) {
    this.docId = docId
    this._version = initialVersion
    this.inflight = []
    this.onSteps = onSteps
  }

  /** Current confirmed document version (sequence number). */
  get version(): number {
    return this._version
  }

  /**
   * Called when the editor produces local steps to send to the server.
   * Returns the WS message payload; caller is responsible for sending it.
   */
  localSteps(steps: unknown[]): { type: 'ot_steps'; doc_id: string; version: number; steps: unknown[] } {
    this.inflight = [...this.inflight, ...steps]
    return {
      type: 'ot_steps',
      doc_id: this.docId,
      version: this._version,
      steps,
    }
  }

  /**
   * Handles incoming WS messages related to OT:
   *  - ot_ack:   server confirmed our steps; advance version.
   *  - ot_steps: remote steps from another client; deliver to editor.
   *  - ot_rebase: version conflict; re-send buffered steps at new base.
   */
  handleMessage(msg: { type: string; steps?: unknown[]; version?: number }): void {
    switch (msg.type) {
      case 'ot_ack':
        // Server confirmed our in-flight steps. Advance version.
        this._version = (msg.version ?? this._version) + 1
        this.inflight = []
        break

      case 'ot_steps':
        // Remote steps from another client — deliver to editor.
        if (msg.steps && msg.steps.length > 0) {
          this.onSteps(msg.steps)
          this._version = msg.version ?? this._version
        }
        break

      case 'ot_rebase':
        // Server rejected our steps due to version mismatch.
        // The server sends the steps we missed; apply them, then re-send ours.
        if (msg.steps && msg.steps.length > 0) {
          this.onSteps(msg.steps)
        }
        this._version = msg.version ?? this._version
        // inflight steps will be re-sent by the caller if needed
        break
    }
  }
}
