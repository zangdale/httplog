var activeSignatures = [];
var burger = document.getElementById("burger");
var filtersCount = document
  .getElementById("filters-count")
  .getElementsByTagName("span")[0];
var interestingFiles = document.getElementById("setting-interesting-files");
var highEntropyStrings = document.getElementById(
  "setting-high-entropy-strings"
);
var notifications = document.getElementById("setting-notifications");
var connectionStats = document.getElementById("connection-status");
var matchesCount = document
  .getElementById("matches-count")
  .getElementsByTagName("span")[0];

const updateStatus = (text, cls) => {
  connectionStats.classList.remove(
    "is-info",
    "is-success",
    "is-warning",
    "is-danger"
  );
  connectionStats.classList.add(cls);
  connectionStats.textContent = text;
};

document.getElementById("filters-clear").addEventListener("click", (event) => {
  activeSignatures = [];
  filtersCount.textContent = "0 filters";

  Array.from(document.querySelectorAll("#signatures li.is-active")).forEach(
    (log) => log.classList.remove("is-active")
  );
  Array.from(document.getElementsByClassName("log")).forEach(
    (log) => (log.style.display = "")
  );
});

burger.addEventListener("click", () => {
  const target = burger.dataset.target;
  const $target = document.getElementById(target);

  burger.classList.toggle("is-active");
  $target.classList.toggle("is-active");
});

interestingFiles.addEventListener("change", (event) => {
  Array.from(document.getElementsByClassName("is-interesting-file")).forEach(
    (log) => {
      log.style.display = event.target.checked ? "" : "none";
    }
  );
});

highEntropyStrings.addEventListener("change", (event) => {
  Array.from(document.getElementsByClassName("is-high-entropy-string")).forEach(
    (log) => {
      log.style.display = event.target.checked ? "" : "none";
    }
  );
});

const notifyFinding = (title, message) => {
  if (Notification.permission === "granted") {
    new Notification(title, {
      icon: "/logo.png",
      body: message,
    });
  }
};

notifications.addEventListener("change", (event) => {
  Notification.requestPermission().then((permission) => {
    if (permission !== "granted") {
      notifications.checked = false;
      notifications.disabled = true;
    }
  });
});

const slugify = (value) =>
  value
    .toLowerCase()
    .replace(/[^a-z0-9 -]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-");

const filterSignature = (signature) => {
  var state = activeSignatures.includes(signature.id);
  signature.classList.toggle("is-active");

  if (!state) activeSignatures.push(signature.id);

  Array.from(document.getElementsByClassName("log")).forEach(
    (log) => (log.style.display = "none")
  );
  activeSignatures.forEach((signatureId) => {
    Array.from(document.getElementsByClassName(signatureId)).forEach((log) => {
      if (state && signatureId == signature.id) return;
      if (
        !interestingFiles.checked &&
        log.classList.contains("is-interesting-file")
      )
        return;
      if (
        !highEntropyStrings.checked &&
        log.classList.contains("is-high-entropy-string")
      )
        return;

      log.style.display = "";
    });
  });

  if (state) {
    activeSignatures.splice(activeSignatures.indexOf(signature.id), 1);
    var anyActive = activeSignatures.length > 0;
    Array.from(
      document.getElementsByClassName(anyActive ? signature.id : "log")
    ).forEach((log) => {
      if (
        !interestingFiles.checked &&
        log.classList.contains("is-interesting-file")
      )
        return;
      if (
        !highEntropyStrings.checked &&
        log.classList.contains("is-high-entropy-string")
      )
        return;

      log.style.display = anyActive ? "none" : "";
    });
  }

  filtersCount.textContent = `${activeSignatures.length} filters`;
};

const sort = (list) => {
  signatures = list.getElementsByTagName("li");
  Array.from(signatures)
    .sort(
      (a, b) =>
        parseInt(
          b.getElementsByClassName("menu-item")[0].getAttribute("data-badge") ||
            0
        ) -
        parseInt(
          a.getElementsByClassName("menu-item")[0].getAttribute("data-badge") ||
            0
        )
    )
    .forEach((li) => list.appendChild(li));
};

// processEvent 解析获取的后端数据
const processEvent = (data) => {
  var eventId = CryptoJS.MD5(data).toString();
  if (document.getElementById(eventId)) return; // duplicate

  var jsonData = JSON.parse(data);
  var sender = slugify(jsonData.sender);
  addTagFunc(sender);

  var matcheCount = data.Matches ? data.Matches.length : 1;
  var sigMenuItem = document
    .getElementById(sender)
    .getElementsByClassName("menu-item")[0];
  sigMenuItem.setAttribute(
    "data-badge",
    parseInt(sigMenuItem.getAttribute("data-badge") || 0) + matcheCount
  );
  // sort(document.getElementById('signatures'));

  var row = document.getElementById("messages").insertRow(0);
  row.classList.add("log", sender);
  row.insertCell(0).innerHTML = `<td class="source">${jsonData.sender}</td>`;
  row.insertCell(1).innerHTML = `<td class="found">${JSON.stringify(
    jsonData.data
  )}</td>`;
  row.insertCell(2).innerHTML = `<td class="signature-name">${JSON.stringify(
    jsonData.data
  )}</td>`;
  row.insertCell(
    3
  ).innerHTML = `<td class="matches">${jsonData.data.time}</td>`;
  row.insertCell(4).innerHTML = `<td class="file-url">4</td>`;
  row.insertCell(5).innerHTML = `<td class="stars">5</td>`;
  row.addEventListener("click", (event) => {});

  matchesCount.textContent = `${
    document.getElementsByClassName("log").length
  } matches`;

  if (!data.Matches) {
    row.classList.add("is-interesting-file");
    if (!interestingFiles.checked) row.style.display = "none";
  }

  if (data.Signature === "High entropy string") {
    row.classList.add("is-high-entropy-string");
    if (!highEntropyStrings.checked) row.style.display = "none";
  }

  if (activeSignatures.length > 0 && !activeSignatures.includes(sigId))
    row.style.display = "none";
  if (notifications.checked)
    notifyFinding(
      data.Signature,
      data.Matches ? data.Matches.join(", ") : data.File
    );
};

// 获取数据的接口地址
const listenForEvents = () => {
  const source = new EventSource("/events");
  source.addEventListener("open", function (event) {
    console.log("eventsource connection open");
  });
  source.addEventListener("error", function (event) {
    if (event.target.readyState === 0) {
      console.log("reconnecting to eventsource");
      updateStatus("Connected", "is-success");
    } else {
      console.error("eventsource error");
    }
  });
  source.addEventListener("message", function (event) {
    updateStatus("Syncing...", "is-info");
    if (document.getElementById("loading"))
      document.getElementById("loading").remove();

    console.log(event.data);
    processEvent(event.data);
  });
};

var STORAGE_KEY = "http-log-senders-v0.1";
// [{"title":"client-tag"},{"title":"client"}]
var senders = JSON.parse(localStorage.getItem(STORAGE_KEY) || "[]");

const saveTagsFunc = function () {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(senders));
};

const addTagFunc = function (title) {
  if (title === "" || title == null || title == undefined) {
    return;
  }

  var findSame = senders.find(function (sender) {
    return sender.title === title;
  });
  if (findSame !== undefined) {
    return;
  }

  senders.push({ title: title });

  var li = document.createElement("li");
  li.id = title;
  li.innerHTML = `<a href="#" class="menu-item" title="${title}">${title}</a>`;
  li.addEventListener("click", (event) => {
    event.preventDefault();
    filterSignature(li);
  });

  document.getElementById("signatures").appendChild(li);

  saveTagsFunc();
};

// 左侧分类标签
const signaturesFunc = () => {
  senders = JSON.parse(localStorage.getItem(STORAGE_KEY) || "[]");

  console.log(senders);

  senders.forEach((sender) => {
    if (JSON.stringify(sender) === "{}") {
      return;
    }

    var li = document.createElement("li");
    li.id = slugify(sender.title);
    li.innerHTML = `<a href="#" class="menu-item" title="${sender.title}">${sender.title}</a>`;
    li.addEventListener("click", (event) => {
      event.preventDefault();
      filterSignature(li);
    });

    document.getElementById("signatures").appendChild(li);
  });

  listenForEvents();

  // window.location.reload();
};

// httplog
document.addEventListener("DOMContentLoaded", function (event) {
  signaturesFunc();
});
