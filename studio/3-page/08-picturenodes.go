// pictureNodes are a picture's page's two: the picture, open, and the note it
// was drawn from, folded, with how to read it; then the page's main ends.
// Filled with the picture, the note's path twice, and the note.
const pictureNodes = `
    <details id="picture" open>
      <summary><h2>Parts and messages</h2></summary>
      <div>
%s
      </div>
    </details>

    <details id="source">
      <summary><h2>Drawn from %s</h2></summary>
      <div>
        <p>
          Each item is a box, and the items under it the boxes in it. An
          item that starts with → is a message, from the box it is under
          to the box it names; its note is what the message carries.
          <code>studio picture %s</code> draws it.
        </p>
        <pre>%s</pre>
      </div>
    </details>
  </main>`
