// Builds a left-sidebar table of contents from the H2/H3 headings that
// appear in the search results (dictionary title / hit heading pairs).
// On narrow screens there is no room for a sidebar, so the CSS in
// default.css falls back to showing this as a normal block placed above
// the results instead.
(function () {
  "use strict";

  function buildTOC() {
    var headings = document.querySelectorAll("div.day > h2, div.section > h3");
    if (headings.length === 0) {
      return;
    }

    var nav = document.createElement("nav");
    nav.id = "toc-sidebar";

    var title = document.createElement("p");
    title.className = "toc-title";
    title.textContent = "目次";
    nav.appendChild(title);

    var topList = document.createElement("ul");
    nav.appendChild(topList);

    var currentSubList = null;

    headings.forEach(function (heading, i) {
      heading.id = heading.id || "toc-" + i;

      var link = document.createElement("a");
      link.href = "#" + heading.id;
      link.textContent = heading.textContent.trim();

      var item = document.createElement("li");
      item.appendChild(link);

      if (heading.tagName === "H2") {
        topList.appendChild(item);
        currentSubList = document.createElement("ul");
        item.appendChild(currentSubList);
      } else if (currentSubList) {
        currentSubList.appendChild(item);
      } else {
        // H3 with no preceding H2 (should not normally happen); keep it
        // visible as a top-level entry rather than dropping it silently.
        topList.appendChild(item);
      }
    });

    insertTOC(nav);
  }

  // Places the TOC immediately above the first search result block so it
  // never ends up inside the query form at the top of the page.
  function insertTOC(nav) {
    var firstDay = document.querySelector("div.day");
    if (!firstDay) {
      return;
    }
    var anchor = firstDay;
    if (
      firstDay.previousElementSibling &&
      firstDay.previousElementSibling.classList.contains("sep")
    ) {
      anchor = firstDay.previousElementSibling;
    }
    anchor.parentNode.insertBefore(nav, anchor);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", buildTOC);
  } else {
    buildTOC();
  }
})();
