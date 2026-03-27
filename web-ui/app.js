// Sucheme Web REPL Application

(function () {
    "use strict";

    var VERSION = "0.1.0";

    // DOM elements
    var loadingEl = document.getElementById("loading");
    var replEl = document.getElementById("repl");
    var errorEl = document.getElementById("error-message");
    var outputEl = document.getElementById("output");
    var inputEl = document.getElementById("input");
    var evalBtn = document.getElementById("eval-btn");

    // Command history
    var history = [];
    var historyIndex = -1;

    // Append a line to the output area
    function appendOutput(text, className) {
        var span = document.createElement("span");
        span.className = className || "";
        span.textContent = text + "\n";
        outputEl.appendChild(span);
        outputEl.scrollTop = outputEl.scrollHeight;
    }

    // Show the welcome message
    function showWelcome() {
        appendOutput("Sucheme v" + VERSION + " - R7RS Scheme Interpreter", "info");
        appendOutput("Type Scheme expressions below and press Ctrl+Enter or click Evaluate.", "info");
        appendOutput("", "info");
    }

    // Evaluate a Scheme expression
    function evaluate(expr) {
        expr = expr.trim();
        if (!expr) {
            return;
        }

        // Add to history
        history.push(expr);
        historyIndex = history.length;

        // Show the input expression
        appendOutput("> " + expr, "prompt");

        // Call the WASM evaluate function if available
        if (typeof window.suchemeEval === "function") {
            try {
                var res = window.suchemeEval(expr);
                if (res && typeof res === "object") {
                    // Structured response: { result, output, error }
                    if (res.output) {
                        // Trim trailing newline to avoid double newline with appendOutput
                        var out = res.output.replace(/\n$/, "");
                        if (out || res.output === "\n") {
                            appendOutput(out, "output");
                        }
                    }
                    if (res.error) {
                        appendOutput("Error: " + res.error, "error");
                    } else if (res.result) {
                        appendOutput(res.result, "result");
                    }
                } else if (res !== undefined && res !== null && res !== "") {
                    // Legacy string response
                    appendOutput(String(res), "result");
                }
            } catch (err) {
                appendOutput("Error: " + err.message, "error");
            }
        } else {
            appendOutput("Error: Interpreter not loaded", "error");
        }
    }

    // Handle the evaluate action
    function handleEvaluate() {
        var expr = inputEl.value;
        evaluate(expr);
        inputEl.value = "";
        inputEl.focus();
    }

    // Navigate command history
    function navigateHistory(direction) {
        if (history.length === 0) {
            return;
        }

        if (direction === "up") {
            if (historyIndex > 0) {
                historyIndex--;
                inputEl.value = history[historyIndex];
            }
        } else if (direction === "down") {
            if (historyIndex < history.length - 1) {
                historyIndex++;
                inputEl.value = history[historyIndex];
            } else {
                historyIndex = history.length;
                inputEl.value = "";
            }
        }
    }

    // Set up keyboard event handlers
    function setupKeyboardHandlers() {
        inputEl.addEventListener("keydown", function (e) {
            // Ctrl+Enter or Cmd+Enter to evaluate
            if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
                e.preventDefault();
                handleEvaluate();
                return;
            }

            // Up arrow for history (only when cursor is at the start)
            if (e.key === "ArrowUp" && inputEl.selectionStart === 0) {
                e.preventDefault();
                navigateHistory("up");
                return;
            }

            // Down arrow for history (only when cursor is at the end)
            if (e.key === "ArrowDown" && inputEl.selectionStart === inputEl.value.length) {
                e.preventDefault();
                navigateHistory("down");
                return;
            }
        });
    }

    // Initialize the WASM module
    async function initWasm() {
        if (typeof Go === "undefined") {
            loadingEl.classList.add("hidden");
            errorEl.classList.remove("hidden");
            return;
        }

        var go = new Go();

        try {
            var result = await WebAssembly.instantiateStreaming(
                fetch("sucheme.wasm"),
                go.importObject
            );
            go.run(result.instance);

            // WASM loaded successfully
            loadingEl.classList.add("hidden");
            replEl.classList.remove("hidden");
            showWelcome();
            inputEl.focus();
        } catch (err) {
            console.error("Failed to load WASM module:", err);
            loadingEl.classList.add("hidden");
            errorEl.classList.remove("hidden");
        }
    }

    // Set up event listeners
    evalBtn.addEventListener("click", handleEvaluate);
    setupKeyboardHandlers();

    // Start loading
    initWasm();
})();
