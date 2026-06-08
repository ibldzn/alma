(function () {
  function readChartData(id) {
    var node = document.getElementById(id);
    if (!node) {
      return null;
    }

    try {
      return JSON.parse(node.textContent || "{}");
    } catch (error) {
      return null;
    }
  }

  function formatCompactRupiah(value) {
    var abs = Math.abs(Number(value) || 0);
    var sign = value < 0 ? "-" : "";
    var unit = "";
    var divisor = 1;

    if (abs >= 1000000000000) {
      unit = " T";
      divisor = 1000000000000;
    } else if (abs >= 1000000000) {
      unit = " M";
      divisor = 1000000000;
    } else if (abs >= 1000000) {
      unit = " jt";
      divisor = 1000000;
    }

    if (unit) {
      return (
        sign +
        "Rp " +
        new Intl.NumberFormat("id-ID", {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2,
        }).format(abs / divisor) +
        unit
      );
    }

    return (
      sign +
      "Rp " +
      new Intl.NumberFormat("id-ID", {
        maximumFractionDigits: 0,
      }).format(abs)
    );
  }

  function formatPercent(value) {
    return (
      new Intl.NumberFormat("id-ID", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value) + "%"
    );
  }

  function createLineChart(canvasId, dataId, valueType) {
    var canvas = document.getElementById(canvasId);
    var data = readChartData(dataId);
    if (
      !canvas ||
      !data ||
      !data.labels ||
      data.labels.length === 0 ||
      !window.Chart
    ) {
      return;
    }

    var colors = ["#0066cc", "#7a7a7a", "#2997ff"];
    data.datasets = (data.datasets || []).map(function (dataset, index) {
      var color = colors[index % colors.length];
      return Object.assign({}, dataset, {
        borderColor: color,
        backgroundColor: color,
        borderWidth: 2,
        pointRadius: 2,
        pointHoverRadius: 4,
        tension: 0.32,
      });
    });

    new Chart(canvas, {
      type: "line",
      data: data,
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: "index",
        },
        plugins: {
          legend: {
            labels: {
              color: "#1d1d1f",
              usePointStyle: true,
              boxWidth: 8,
              boxHeight: 8,
            },
          },
          tooltip: {
            callbacks: {
              label: function (context) {
                var label = context.dataset.label
                  ? context.dataset.label + ": "
                  : "";
                var value = context.parsed.y;
                return (
                  label +
                  (valueType === "percent"
                    ? formatPercent(value)
                    : formatCompactRupiah(value))
                );
              },
            },
          },
        },
        scales: {
          x: {
            grid: {
              color: "rgba(0, 0, 0, 0.06)",
            },
            ticks: {
              color: "#7a7a7a",
              maxRotation: 0,
              autoSkip: true,
            },
          },
          y: {
            beginAtZero: false,
            grid: {
              color: "rgba(0, 0, 0, 0.06)",
            },
            ticks: {
              color: "#7a7a7a",
              callback: function (value) {
                return valueType === "percent"
                  ? formatPercent(value)
                  : formatCompactRupiah(value);
              },
            },
          },
        },
      },
    });
  }

  function syncDateFilterInputs() {
    var rangeInput = document.querySelector('select[name="range"]');
    var dateInputs = document.querySelectorAll(
      'input[name="start_date"], input[name="end_date"]',
    );
    if (!rangeInput || dateInputs.length === 0) {
      return;
    }

    function updateDateInputs() {
      var customRangeSelected = rangeInput.value.toLowerCase() === "custom";
      Array.prototype.forEach.call(dateInputs, function (input) {
        input.disabled = !customRangeSelected;
      });
    }

    rangeInput.addEventListener("change", updateDateInputs);
    updateDateInputs();
  }

  function syncFilterMenu() {
    var subNav = document.querySelector(".sub-nav");
    var toggle = document.querySelector(".filter-menu-toggle");
    if (!subNav || !toggle) {
      return;
    }

    var targetId = toggle.getAttribute("aria-controls");
    var filters = targetId ? document.getElementById(targetId) : null;
    if (!filters) {
      return;
    }

    var mobileQuery = window.matchMedia("(max-width: 640px)");

    function setOpen(open) {
      var shouldOpen = Boolean(open && mobileQuery.matches);
      subNav.classList.toggle("is-open", shouldOpen);
      toggle.setAttribute("aria-expanded", shouldOpen ? "true" : "false");
    }

    toggle.addEventListener("click", function () {
      setOpen(toggle.getAttribute("aria-expanded") !== "true");
    });

    filters.addEventListener("submit", function () {
      setOpen(false);
    });

    document.addEventListener("keydown", function (event) {
      if (
        event.key === "Escape" &&
        toggle.getAttribute("aria-expanded") === "true"
      ) {
        setOpen(false);
        toggle.focus();
      }
    });

    if (typeof mobileQuery.addEventListener === "function") {
      mobileQuery.addEventListener("change", function () {
        setOpen(false);
      });
    } else if (typeof mobileQuery.addListener === "function") {
      mobileQuery.addListener(function () {
        setOpen(false);
      });
    }
  }

  syncFilterMenu();
  syncDateFilterInputs();
  createLineChart(
    "historical-deposits-chart",
    "historical-deposits-data",
    "money",
  );
  createLineChart("historical-ldr-chart", "historical-ldr-data", "percent");
})();
