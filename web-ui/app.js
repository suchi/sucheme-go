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
    var clearOutputBtn = document.getElementById("clear-output-btn");
    var copyOutputBtn = document.getElementById("copy-output-btn");
    var keybindSelect = document.getElementById("keybind-select");
    var themeToggleBtn = document.getElementById("theme-toggle-btn");
    var fontIncreaseBtn = document.getElementById("font-increase-btn");
    var fontDecreaseBtn = document.getElementById("font-decrease-btn");
    var viModeIndicator = document.getElementById("vi-mode-indicator");
    var highlightOverlay = document.getElementById("highlight-overlay");

    // =========================================================================
    // Persistent History (localStorage)
    // =========================================================================
    var HISTORY_KEY = "sucheme_history";
    var MAX_HISTORY = 100;

    function loadHistory() {
        try {
            var data = localStorage.getItem(HISTORY_KEY);
            if (data) {
                var parsed = JSON.parse(data);
                if (Array.isArray(parsed)) {
                    return parsed.slice(-MAX_HISTORY);
                }
            }
        } catch (e) {
            // ignore
        }
        return [];
    }

    function saveHistory(hist) {
        try {
            localStorage.setItem(HISTORY_KEY, JSON.stringify(hist.slice(-MAX_HISTORY)));
        } catch (e) {
            // ignore
        }
    }

    var history = loadHistory();
    var historyIndex = history.length;

    function addToHistory(expr) {
        history.push(expr);
        // Trim in-memory array to MAX_HISTORY to prevent unbounded growth
        if (history.length > MAX_HISTORY) {
            history = history.slice(-MAX_HISTORY);
        }
        historyIndex = history.length;
        saveHistory(history);
    }

    // =========================================================================
    // Theme Management
    // =========================================================================
    var THEME_KEY = "sucheme_theme";

    function getStoredTheme() {
        try {
            return localStorage.getItem(THEME_KEY) || "dark";
        } catch (e) {
            return "dark";
        }
    }

    function applyTheme(theme) {
        if (theme === "light") {
            document.body.classList.add("light-theme");
            themeToggleBtn.textContent = "Dark";
        } else {
            document.body.classList.remove("light-theme");
            themeToggleBtn.textContent = "Light";
        }
        try {
            localStorage.setItem(THEME_KEY, theme);
        } catch (e) {
            // ignore
        }
    }

    function toggleTheme() {
        var current = document.body.classList.contains("light-theme") ? "light" : "dark";
        applyTheme(current === "light" ? "dark" : "light");
    }

    // =========================================================================
    // Font Size Management
    // =========================================================================
    var FONT_SIZE_KEY = "sucheme_font_size";
    var MIN_FONT_SIZE = 10;
    var MAX_FONT_SIZE = 24;
    var DEFAULT_FONT_SIZE = 14;

    function getStoredFontSize() {
        try {
            var size = parseInt(localStorage.getItem(FONT_SIZE_KEY), 10);
            if (!isNaN(size) && size >= MIN_FONT_SIZE && size <= MAX_FONT_SIZE) {
                return size;
            }
        } catch (e) {
            // ignore
        }
        return DEFAULT_FONT_SIZE;
    }

    var currentFontSize = getStoredFontSize();

    function applyFontSize(size) {
        size = Math.max(MIN_FONT_SIZE, Math.min(MAX_FONT_SIZE, size));
        currentFontSize = size;
        outputEl.style.fontSize = size + "px";
        inputEl.style.fontSize = size + "px";
        if (highlightOverlay) {
            highlightOverlay.style.fontSize = size + "px";
        }
        try {
            localStorage.setItem(FONT_SIZE_KEY, String(size));
        } catch (e) {
            // ignore
        }
    }

    function increaseFontSize() {
        applyFontSize(currentFontSize + 1);
    }

    function decreaseFontSize() {
        applyFontSize(currentFontSize - 1);
    }

    // =========================================================================
    // Keybinding Modes
    // =========================================================================
    var KEYBIND_KEY = "sucheme_keybind";
    var keybindMode = "emacs"; // "emacs" | "vi" | "none"
    var viMode = "insert"; // "insert" | "normal"

    function getStoredKeybind() {
        try {
            var mode = localStorage.getItem(KEYBIND_KEY);
            if (mode === "emacs" || mode === "vi" || mode === "none") {
                return mode;
            }
        } catch (e) {
            // ignore
        }
        return "emacs";
    }

    function setKeybindMode(mode) {
        keybindMode = mode;
        if (mode === "vi") {
            viMode = "insert";
            viModeIndicator.classList.remove("hidden");
            viModeIndicator.textContent = "-- INSERT --";
        } else {
            viModeIndicator.classList.add("hidden");
        }
        try {
            localStorage.setItem(KEYBIND_KEY, mode);
        } catch (e) {
            // ignore
        }
    }

    function setViMode(mode) {
        viMode = mode;
        if (mode === "normal") {
            viModeIndicator.textContent = "-- NORMAL --";
        } else {
            viModeIndicator.textContent = "-- INSERT --";
        }
    }

    // =========================================================================
    // Syntax Highlighting
    // =========================================================================
    var SCHEME_KEYWORDS = [
        "define", "lambda", "if", "cond", "else", "let", "let\\*", "letrec",
        "begin", "set!", "quote", "and", "or", "not", "when", "unless",
        "do", "case", "define-syntax", "syntax-rules", "define-record-type",
        "import", "export", "library", "values", "call-with-current-continuation",
        "call/cc", "apply", "map", "for-each", "display", "newline", "write",
        "car", "cdr", "cons", "list", "null\\?", "pair\\?", "number\\?",
        "string\\?", "symbol\\?", "boolean\\?", "vector\\?", "char\\?",
        "eq\\?", "eqv\\?", "equal\\?"
    ];

    var keywordPattern = new RegExp(
        "\\b(" + SCHEME_KEYWORDS.join("|") + ")\\b", "g"
    );

    // Pattern for strings, comments, numbers, and booleans
    var stringPattern = /"(?:[^"\\]|\\.)*"/g;
    var commentPattern = /;[^\n]*/g;
    var numberPattern;
    try {
        // Use lookbehind where supported for more precise number matching
        numberPattern = new RegExp(
            "(?<![a-zA-Z!$%&*+\\-./:<=>?@^_~#])" +
            "([+-]?\\d+(?:\\.\\d+)?(?:[eE][+-]?\\d+)?|[+-]?\\d+/\\d+)" +
            "(?![a-zA-Z!$%&*+\\-./:<=>?@^_~#])",
            "g"
        );
    } catch (e) {
        // Fallback for environments without lookbehind support
        numberPattern = /[+-]?(?:\d+(?:\.\d+)?(?:[eE][+-]?\d+)?|\d+\/\d+)/g;
    }
    var booleanPattern = /#[tf]\b/g;

    function escapeHtml(text) {
        return text
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;");
    }

    function highlightScheme(text) {
        if (!text) return "";
        // Tokenize and highlight
        var escaped = escapeHtml(text);

        // Mark regions that should not be further highlighted
        var regions = [];

        // Highlight strings first
        escaped = escaped.replace(stringPattern, function (match, offset) {
            var placeholder = "\x00STR" + regions.length + "\x00";
            regions.push('<span class="hl-string">' + match + '</span>');
            return placeholder;
        });

        // Highlight comments
        escaped = escaped.replace(commentPattern, function (match, offset) {
            var placeholder = "\x00CMT" + regions.length + "\x00";
            regions.push('<span class="hl-comment">' + match + '</span>');
            return placeholder;
        });

        // Highlight booleans
        escaped = escaped.replace(booleanPattern, function (match) {
            var placeholder = "\x00BOOL" + regions.length + "\x00";
            regions.push('<span class="hl-boolean">' + match + '</span>');
            return placeholder;
        });

        // Highlight numbers
        escaped = escaped.replace(numberPattern, function (match) {
            var placeholder = "\x00NUM" + regions.length + "\x00";
            regions.push('<span class="hl-number">' + match + '</span>');
            return placeholder;
        });

        // Highlight keywords
        escaped = escaped.replace(keywordPattern, function (match) {
            return '<span class="hl-keyword">' + match + '</span>';
        });

        // Restore protected regions
        for (var i = 0; i < regions.length; i++) {
            escaped = escaped.replace(new RegExp("\x00(?:STR|CMT|BOOL|NUM)" + i + "\x00"), regions[i]);
        }

        return escaped;
    }

    function updateHighlightOverlay() {
        if (!highlightOverlay) return;
        var text = inputEl.value;
        // Add a trailing newline so the overlay height matches the textarea
        highlightOverlay.innerHTML = highlightScheme(text) + "\n";
        // Sync scroll
        highlightOverlay.scrollTop = inputEl.scrollTop;
        highlightOverlay.scrollLeft = inputEl.scrollLeft;
    }

    // =========================================================================
    // Output Management
    // =========================================================================

    // Append a line to the output area
    function appendOutput(text, className) {
        var span = document.createElement("span");
        span.className = className || "";
        span.textContent = text + "\n";
        outputEl.appendChild(span);
        outputEl.scrollTop = outputEl.scrollHeight;
    }

    function clearOutput() {
        outputEl.innerHTML = "";
    }

    function copyOutput() {
        var text = outputEl.textContent || outputEl.innerText;
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function () {
                var orig = copyOutputBtn.textContent;
                copyOutputBtn.textContent = "Copied!";
                setTimeout(function () {
                    copyOutputBtn.textContent = orig;
                }, 1500);
            }).catch(function () {
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    }

    function fallbackCopy(text) {
        var ta = document.createElement("textarea");
        ta.value = text;
        ta.style.position = "fixed";
        ta.style.left = "-9999px";
        document.body.appendChild(ta);
        ta.select();
        try {
            document.execCommand("copy");
            var orig = copyOutputBtn.textContent;
            copyOutputBtn.textContent = "Copied!";
            setTimeout(function () {
                copyOutputBtn.textContent = orig;
            }, 1500);
        } catch (e) {
            // ignore
        }
        document.body.removeChild(ta);
    }

    // =========================================================================
    // Help and Examples
    // =========================================================================

    var HELP_TEXT = [
        "=== Sucheme Help ===",
        "",
        "Built-in Special Forms:",
        "  define, lambda, if, cond, else, case, and, or, not",
        "  let, let*, letrec, begin, set!, quote, when, unless, do",
        "  define-syntax, syntax-rules, define-record-type",
        "  import, export, library",
        "",
        "Built-in Procedures:",
        "  Arithmetic: +, -, *, /, =, <, >, <=, >=, zero?, positive?, negative?",
        "  Lists:      car, cdr, cons, list, append, reverse, length, map, for-each, filter",
        "  Pairs:      pair?, null?, list?",
        "  Type:       number?, string?, symbol?, boolean?, char?, vector?, procedure?",
        "  Equality:   eq?, eqv?, equal?",
        "  Strings:    string-length, string-ref, string-append, substring",
        "  I/O:        display, newline, write",
        "  Control:    apply, call/cc, values, call-with-values",
        "  Misc:       not, abs, min, max, modulo, remainder, expt, sqrt",
        "",
        "REPL Commands:",
        "  :help       Show this help",
        "  :examples   Show example expressions",
        "  :clear      Clear output",
        "",
        "Keyboard Shortcuts:",
        "  Ctrl+Enter  Evaluate expression",
        "  Up/Down     Navigate history",
        "  Ctrl++/-    Adjust font size",
        ""
    ];

    var EXAMPLES_TEXT = [
        "=== Example Expressions ===",
        "",
        ";; Basic arithmetic",
        "(+ 1 2 3)                    ; => 6",
        "(* 2 (+ 3 4))               ; => 14",
        "",
        ";; Define a variable",
        "(define x 42)",
        "x                            ; => 42",
        "",
        ";; Define a function",
        "(define (square n) (* n n))",
        "(square 5)                   ; => 25",
        "",
        ";; Lambda",
        "((lambda (x y) (+ x y)) 3 4) ; => 7",
        "",
        ";; List operations",
        "(list 1 2 3)                 ; => (1 2 3)",
        "(car '(1 2 3))              ; => 1",
        "(cdr '(1 2 3))              ; => (2 3)",
        "(map (lambda (x) (* x x)) '(1 2 3 4 5))",
        "",
        ";; Conditional",
        "(if (> 3 2) \"yes\" \"no\")      ; => \"yes\"",
        "",
        ";; Recursion: factorial",
        "(define (fact n)",
        "  (if (<= n 1) 1 (* n (fact (- n 1)))))",
        "(fact 10)                    ; => 3628800",
        "",
        ";; Higher-order function",
        "(define (compose f g)",
        "  (lambda (x) (f (g x))))",
        "((compose square (lambda (x) (+ x 1))) 4) ; => 25",
        ""
    ];

    function showHelp() {
        HELP_TEXT.forEach(function (line) {
            appendOutput(line, "info");
        });
    }

    function showExamples() {
        EXAMPLES_TEXT.forEach(function (line) {
            appendOutput(line, "info");
        });
    }

    // =========================================================================
    // Show the welcome message
    // =========================================================================
    function showWelcome() {
        appendOutput("Sucheme v" + VERSION + " - R7RS Scheme Interpreter", "info");
        appendOutput("Type Scheme expressions below and press Ctrl+Enter or click Evaluate.", "info");
        appendOutput("Type :help for commands, :examples for sample code.", "info");
        appendOutput("", "info");
    }

    // =========================================================================
    // Evaluate a Scheme expression
    // =========================================================================
    function evaluate(expr) {
        expr = expr.trim();
        if (!expr) {
            return;
        }

        // Handle REPL meta-commands
        if (expr === ":help") {
            addToHistory(expr);
            appendOutput("> " + expr, "prompt");
            showHelp();
            return;
        }
        if (expr === ":examples") {
            addToHistory(expr);
            appendOutput("> " + expr, "prompt");
            showExamples();
            return;
        }
        if (expr === ":clear") {
            addToHistory(expr);
            clearOutput();
            return;
        }

        // Add to history
        addToHistory(expr);

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
        updateHighlightOverlay();
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
        updateHighlightOverlay();
    }

    // =========================================================================
    // Vi Mode Handling
    // =========================================================================
    function handleViNormalKey(e) {
        switch (e.key) {
        case "i":
            e.preventDefault();
            setViMode("insert");
            return true;
        case "a":
            e.preventDefault();
            setViMode("insert");
            // move cursor one right
            if (inputEl.selectionStart < inputEl.value.length) {
                inputEl.selectionStart = inputEl.selectionEnd = inputEl.selectionStart + 1;
            }
            return true;
        case "I":
            e.preventDefault();
            setViMode("insert");
            inputEl.selectionStart = inputEl.selectionEnd = 0;
            return true;
        case "A":
            e.preventDefault();
            setViMode("insert");
            inputEl.selectionStart = inputEl.selectionEnd = inputEl.value.length;
            return true;
        case "h":
            e.preventDefault();
            if (inputEl.selectionStart > 0) {
                inputEl.selectionStart = inputEl.selectionEnd = inputEl.selectionStart - 1;
            }
            return true;
        case "l":
            e.preventDefault();
            if (inputEl.selectionStart < inputEl.value.length) {
                inputEl.selectionStart = inputEl.selectionEnd = inputEl.selectionStart + 1;
            }
            return true;
        case "j":
            e.preventDefault();
            navigateHistory("down");
            return true;
        case "k":
            e.preventDefault();
            navigateHistory("up");
            return true;
        case "0":
            e.preventDefault();
            inputEl.selectionStart = inputEl.selectionEnd = 0;
            return true;
        case "$":
            e.preventDefault();
            inputEl.selectionStart = inputEl.selectionEnd = inputEl.value.length;
            return true;
        case "x":
            e.preventDefault();
            if (inputEl.selectionStart < inputEl.value.length) {
                var pos = inputEl.selectionStart;
                inputEl.value = inputEl.value.slice(0, pos) + inputEl.value.slice(pos + 1);
                inputEl.selectionStart = inputEl.selectionEnd = pos;
                updateHighlightOverlay();
            }
            return true;
        case "d":
            // wait for second 'd' -- use a simple flag
            if (handleViNormalKey._pendingD) {
                e.preventDefault();
                handleViNormalKey._pendingD = false;
                inputEl.value = "";
                updateHighlightOverlay();
                return true;
            } else {
                e.preventDefault();
                handleViNormalKey._pendingD = true;
                setTimeout(function () { handleViNormalKey._pendingD = false; }, 500);
                return true;
            }
        default:
            // In normal mode, prevent most typing
            if (!e.ctrlKey && !e.metaKey && !e.altKey && e.key.length === 1) {
                e.preventDefault();
                return true;
            }
            return false;
        }
    }
    handleViNormalKey._pendingD = false;

    // =========================================================================
    // Emacs Keybinding Handling
    // =========================================================================
    function handleEmacsKey(e) {
        if (!e.ctrlKey) return false;

        switch (e.key) {
        case "a":
            // Move to beginning of line
            e.preventDefault();
            inputEl.selectionStart = inputEl.selectionEnd = 0;
            return true;
        case "e":
            // Move to end of line
            e.preventDefault();
            inputEl.selectionStart = inputEl.selectionEnd = inputEl.value.length;
            return true;
        case "k":
            // Kill to end of line
            e.preventDefault();
            inputEl.value = inputEl.value.slice(0, inputEl.selectionStart);
            updateHighlightOverlay();
            return true;
        case "d":
            // Delete character forward
            e.preventDefault();
            if (inputEl.selectionStart < inputEl.value.length) {
                var p = inputEl.selectionStart;
                inputEl.value = inputEl.value.slice(0, p) + inputEl.value.slice(p + 1);
                inputEl.selectionStart = inputEl.selectionEnd = p;
                updateHighlightOverlay();
            }
            return true;
        case "f":
            // Move forward one character
            e.preventDefault();
            if (inputEl.selectionStart < inputEl.value.length) {
                inputEl.selectionStart = inputEl.selectionEnd = inputEl.selectionStart + 1;
            }
            return true;
        case "b":
            // Move backward one character
            e.preventDefault();
            if (inputEl.selectionStart > 0) {
                inputEl.selectionStart = inputEl.selectionEnd = inputEl.selectionStart - 1;
            }
            return true;
        default:
            return false;
        }
    }

    // =========================================================================
    // Set up keyboard event handlers
    // =========================================================================
    function setupKeyboardHandlers() {
        inputEl.addEventListener("keydown", function (e) {
            // Ctrl+Enter or Cmd+Enter to evaluate
            if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
                e.preventDefault();
                handleEvaluate();
                return;
            }

            // Vi mode: Escape to enter normal mode
            if (keybindMode === "vi") {
                if (e.key === "Escape") {
                    e.preventDefault();
                    setViMode("normal");
                    return;
                }
                if (viMode === "normal") {
                    if (handleViNormalKey(e)) return;
                }
            }

            // Emacs keybindings
            if (keybindMode === "emacs") {
                if (handleEmacsKey(e)) return;
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

        // Sync highlight overlay on input
        inputEl.addEventListener("input", updateHighlightOverlay);
        inputEl.addEventListener("scroll", function () {
            if (highlightOverlay) {
                highlightOverlay.scrollTop = inputEl.scrollTop;
                highlightOverlay.scrollLeft = inputEl.scrollLeft;
            }
        });

        // Global font size shortcuts
        document.addEventListener("keydown", function (e) {
            if ((e.ctrlKey || e.metaKey) && (e.key === "+" || e.key === "=")) {
                e.preventDefault();
                increaseFontSize();
            }
            if ((e.ctrlKey || e.metaKey) && e.key === "-") {
                e.preventDefault();
                decreaseFontSize();
            }
        });
    }

    // =========================================================================
    // Initialize the WASM module
    // =========================================================================
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

    // =========================================================================
    // Set up event listeners
    // =========================================================================
    evalBtn.addEventListener("click", handleEvaluate);
    clearOutputBtn.addEventListener("click", clearOutput);
    copyOutputBtn.addEventListener("click", copyOutput);

    themeToggleBtn.addEventListener("click", toggleTheme);
    fontIncreaseBtn.addEventListener("click", increaseFontSize);
    fontDecreaseBtn.addEventListener("click", decreaseFontSize);

    keybindSelect.addEventListener("change", function () {
        setKeybindMode(keybindSelect.value);
    });

    setupKeyboardHandlers();

    // Apply stored settings
    applyTheme(getStoredTheme());
    applyFontSize(getStoredFontSize());
    var storedKeybind = getStoredKeybind();
    keybindSelect.value = storedKeybind;
    setKeybindMode(storedKeybind);

    // Start loading
    initWasm();
})();
