export function clearNode(node) {
    node.replaceChildren();
}

export function el(tagName, attrs = {}, ...children) {
    const node = document.createElement(tagName);

    for (const [key, value] of Object.entries(attrs)) {
        if (value == null) {
            continue;
        }

        if (key === "className") {
            node.className = value;
            continue;
        }

        if (key === "text") {
            node.textContent = value;
            continue;
        }

        if (key === "html") {
            node.innerHTML = value;
            continue;
        }

        if (key === "dataset") {
            Object.assign(node.dataset, value);
            continue;
        }

        if (key.startsWith("on") && typeof value === "function") {
            node.addEventListener(key.slice(2).toLowerCase(), value);
            continue;
        }

        node.setAttribute(key, value);
    }

    for (const child of children.flat()) {
        if (child == null || child === false) {
            continue;
        }

        node.append(child.nodeType ? child : document.createTextNode(String(child)));
    }

    return node;
}

export function field({ label, name, value = "", type = "text", required = true, autocomplete, placeholder }) {
    const input = el("input", {
        name,
        type,
        value,
        required,
        autocomplete,
        placeholder,
    });

    return el("label", { className: "field" }, el("span", { text: label }), input);
}

export function textareaField({ label, name, value = "", required = true, placeholder }) {
    const input = el("textarea", {
        name,
        required,
        placeholder,
    });
    input.value = value;
    return el("label", { className: "field" }, el("span", { text: label }), input);
}

export function selectField({ label, name, value = "", required = true, options = [] }) {
    const input = el("select", {
        name,
        required,
    });

    for (const option of options) {
        input.append(el("option", { value: option.value, text: option.label }));
    }

    input.value = value;

    return el("label", { className: "field" }, el("span", { text: label }), input);
}

export function flashMessage(flash) {
    if (!flash) {
        return null;
    }

    return el("div", {
        className: `flash flash-${flash.kind || "info"}`,
        text: flash.message,
    });
}

export function authPage({ mode, onSignIn, onSignUp, onNavigateAuth, flash, metaText }) {
    const isSignUp = mode === "sign-up";
    const form = el("form", { className: "card form-card auth-card" });

    const flashNode = flashMessage(flash);

    form.append(
        el("h2", { text: isSignUp ? "Sign up" : "Sign in" }),
        ...(flashNode ? [flashNode] : []),
        field({ label: "Username", name: "username", autocomplete: "username", required: true }),
        field({
            label: "Password",
            name: "password",
            type: "password",
            autocomplete: isSignUp ? "new-password" : "current-password",
            required: true,
        })
    );

    if (isSignUp) {
        form.append(selectField({
            label: "Language code",
            name: "languageCode",
            value: "en-US",
            required: true,
            options: [
                { value: "en-US", label: "en-US" },
                { value: "fr-FR", label: "fr-FR" },
            ],
        }));
    }

    form.append(
        el("button", {
            type: "submit",
            className: "primary",
            text: isSignUp ? "Create account" : "Sign in",
        }),
        el("div", { className: "auth-switch" },
            el("span", { className: "muted", text: isSignUp ? "Already have an account?" : "Need an account?" }),
            el("button", {
                type: "button",
                className: "text-link",
                text: isSignUp ? "Sign in" : "Sign up",
                onClick: () => onNavigateAuth(isSignUp ? "/sign-in" : "/sign-up"),
            })
        )
    );

    form.addEventListener("submit", async (event) => {
        event.preventDefault();
        if (isSignUp) {
            await onSignUp(form);
            return;
        }
        await onSignIn(form);
    });

    return el("main", { className: "auth-page" },
        el("div", { className: "auth-stack" },
            form,
            el("footer", { className: "auth-footer" },
                el("div", { className: "meta-label", text: "Version" }),
                el("div", { className: "meta-value", text: metaText || "Loading..." })
            )
        )
    );
}

function sidebarButton({ label, active, onClick }) {
    return el("button", {
        type: "button",
        className: active ? "sidebar-button active" : "sidebar-button",
        onClick,
        text: label,
    });
}

export function shell({ route, metaText, onNavigate, onSignOut, content }) {
    return el("main", { className: "shell" },
        el("aside", { className: "sidebar" },
            el("div", { className: "sidebar-top" },
                el("div", { className: "brand" },
                    el("div", { className: "brand-mark", text: "G" }),
                    el("div", {},
                        el("strong", { text: "Gostarter" }),
                        el("p", { className: "muted", text: "Full example" })
                    )
                ),
                el("nav", { className: "sidebar-nav", "aria-label": "Primary" },
                    sidebarButton({ label: "Notes", active: route === "/notes", onClick: () => onNavigate("/notes") }),
                    sidebarButton({ label: "Account", active: route === "/account", onClick: () => onNavigate("/account") }),
                    el("button", { type: "button", className: "sidebar-button", text: "Sign out", onClick: onSignOut })
                )
            ),
            el("footer", { className: "sidebar-footer" },
                el("div", { className: "meta-label", text: "Version" }),
                el("div", { className: "meta-value", text: metaText || "Loading..." })
            )
        ),
        el("section", { className: "content" }, content)
    );
}

function noteCard({ note, onSave, onDelete }) {
    const body = el("textarea", { className: "note-body" });
    body.value = note.body || "";

    return el("article", { className: "card note-card" },
        el("div", { className: "note-head" },
            el("div", {},
                el("strong", { text: "Note" }),
                el("div", { className: "muted", text: note.id })
            ),
            el("div", { className: "note-meta" },
                el("span", { text: `Created ${note.createdAt || "unknown"}` }),
                el("span", { text: `Updated ${note.updatedAt || "unknown"}` })
            )
        ),
        el("label", { className: "field" }, el("span", { text: "Body" }), body),
        el("div", { className: "button-row" },
            el("button", { type: "button", className: "primary", text: "Save", onClick: () => onSave(note.id, body.value) }),
            el("button", { type: "button", className: "danger", text: "Delete", onClick: () => onDelete(note.id) })
        )
    );
}

export function notesPage({ notes, onCreateNote, onRefreshNotes, onSaveNote, onDeleteNote }) {
    const list = notes.length
        ? el("div", { className: "notes-list" }, notes.map((note) => noteCard({ note, onSave: onSaveNote, onDelete: onDeleteNote })))
        : el("div", { className: "empty-state", text: "No notes yet. Create one to start." });

    return el("div", {},
        el("div", { className: "page-head" },
            el("div", {},
                el("p", { className: "eyebrow", text: "Notes" }),
                el("h1", { text: "Your notes" })
            ),
            el("div", { className: "button-row" },
                el("button", { type: "button", className: "primary", text: "Create note", onClick: onCreateNote }),
                el("button", { type: "button", text: "Refresh", onClick: onRefreshNotes })
            )
        ),
        list
    );
}

export function accountPage({ account, onUpdateUsername, onUpdatePassword, onUpdateLanguage, onDeleteAccount }) {
    const usernameForm = el("form", { className: "card form-card" });
    usernameForm.append(
        el("h2", { text: "Update username" }),
        field({ label: "New username", name: "newUsername", value: account?.username || "", required: true }),
        el("button", { type: "submit", className: "primary", text: "Save username" })
    );
    usernameForm.addEventListener("submit", async (event) => {
        event.preventDefault();
        await onUpdateUsername(usernameForm);
    });

    const passwordForm = el("form", { className: "card form-card" });
    passwordForm.append(
        el("h2", { text: "Update password" }),
        field({ label: "Old password", name: "oldPassword", type: "password", autocomplete: "current-password", required: true }),
        field({ label: "New password", name: "newPassword", type: "password", autocomplete: "new-password", required: true }),
        el("button", { type: "submit", className: "primary", text: "Save password" })
    );
    passwordForm.addEventListener("submit", async (event) => {
        event.preventDefault();
        await onUpdatePassword(passwordForm);
    });

    const languageForm = el("form", { className: "card form-card" });
    languageForm.append(
        el("h2", { text: "Update language" }),
        selectField({
            label: "Language code",
            name: "languageCode",
            value: account?.languageCode || "en-US",
            required: true,
            options: [
                { value: "en-US", label: "en-US" },
                { value: "fr-FR", label: "fr-FR" },
            ],
        }),
        el("button", { type: "submit", className: "primary", text: "Save language" })
    );
    languageForm.addEventListener("submit", async (event) => {
        event.preventDefault();
        await onUpdateLanguage(languageForm);
    });

    return el("div", {},
        el("div", { className: "page-head" },
            el("div", {},
                el("p", { className: "eyebrow", text: "Account" }),
                el("h1", { text: account ? account.username : "Account" }),
                el("p", { className: "lead", text: "Update your profile settings and account credentials here." })
            ),
            el("button", { type: "button", className: "danger", text: "Delete account", onClick: onDeleteAccount })
        ),
        el("div", { className: "card account-summary" },
            el("div", {}, el("span", { className: "muted", text: "Role" }), el("strong", { text: account?.role || "-" })),
            el("div", {}, el("span", { className: "muted", text: "Language" }), el("strong", { text: account?.languageCode || "-" })),
            el("div", {}, el("span", { className: "muted", text: "Created" }), el("strong", { text: account?.createdAt || "-" }))
        ),
        el("div", { className: "form-grid" }, usernameForm, passwordForm, languageForm)
    );
}