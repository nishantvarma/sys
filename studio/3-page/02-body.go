// body is the page's main: the title, the lead, the frame, and the note
// it was laid out from, folded.
const body = `<main>
    <h1>%s</h1>
    <div class="lead">
      <p>%s</p>
    </div>

    <details id="frame" open>
      <summary><h2>The frame</h2></summary>
      <div>
        <div class="frame">
%s        </div>
      </div>
    </details>

    <details id="note">
      <summary><h2>Laid out from %s</h2></summary>
      <div>
        <pre>%s</pre>
      </div>
    </details>
  </main>`
