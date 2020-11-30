window.onload = function () {
  // do nothing if browser doesn't support the copy command
  if (!document.queryCommandSupported("copy")) {
    return;
  }

  // add copy button to each code block
  const higlights = document.getElementsByClassName("highlight");
  for (let i = 0; i < higlights.length; i++) {
    addButton(higlights[i]);
  }
};

function addButton(container) {
  const btn = document.createElement("button");
  btn.className = "copy-button";
  btn.textContent = "📋";

  btn.addEventListener("click", function () {
    const selection = window.getSelection();

    // deselect anything selected
    selection.removeAllRanges();

    // select the contents of our code block
    const range = document.createRange();
    range.selectNodeContents(container.firstElementChild);
    selection.addRange(range);

    // copy to clipboard and then remove selection
    document.execCommand("copy");
    selection.removeAllRanges();
  });

  container.appendChild(btn);
}
