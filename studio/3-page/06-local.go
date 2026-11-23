// local is the frame's look on the page, under the skeleton's local mark:
// slots as boxes that fill the frame, in the page's own colours.
const local = `
    .frame { display: flex; height: 60vh; min-height: 320px; }
    .frame .down, .frame .across {
      display: flex; flex: 1; gap: 8px; min-width: 0; min-height: 0;
    }
    .frame .down { flex-direction: column; }
    .frame .slot { display: flex; min-width: 0; min-height: 0; }
    .frame .pane {
      flex: 1; overflow: auto; padding: 10px 12px;
      border: 1px solid var(--line); border-radius: 8px;
      background: var(--panel);
    }
    .frame .name { margin: 0 0 6px; font-weight: 600; }
    .frame .waits { margin: 0; color: var(--dim); }
`
