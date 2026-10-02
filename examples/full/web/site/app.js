import { clearNode, authPage, shell, notesPage, accountPage, usersPage } from "./ui.js";

const root = document.getElementById("app");

// Ordered from the highest to the lowest role, same as the role manager of the
// example server. Routes gated by a role are prefixed with it, so /moderator is
// reachable by moderators and owners, and /owner by owners only.
const roles = ["owner", "moderator", "user"];

const isPrivileged = (account) => account?.role === "owner" || account?.role === "moderator";

const state = {
    signedIn: false,
    csrfToken: "",
    meta: null,
    account: null,
    notes: [],
    accounts: [],
    route: "/sign-in",
    flash: null,
};

const normalizePath = (path) => {
    const cleaned = path.replace(/\/+$/, "");
    return cleaned === "" ? "/" : cleaned;
};

const currentPath = () => normalizePath(window.location.pathname);

function setFlash(message, kind = "info") {
    state.flash = { message, kind };
    render();
}

function clearFlash() {
    state.flash = null;
}

function navigate(path, replace = false) {
    const next = normalizePath(path);
    if (replace) {
        window.history.replaceState({}, "", next);
    } else if (next !== currentPath()) {
        window.history.pushState({}, "", next);
    }
    state.route = next;
}

async function apiFetch(path, options = {}) {
    const headers = new Headers(options.headers || {});
    const method = (options.method || "GET").toUpperCase();

    if (options.body && !headers.has("Content-Type")) {
        headers.set("Content-Type", "application/json");
    }

    if (["POST", "PUT", "PATCH", "DELETE"].includes(method) && state.csrfToken && !headers.has("X-CSRF-Token")) {
        headers.set("X-CSRF-Token", state.csrfToken);
    }

    const response = await fetch(`/api${path}`, {
        ...options,
        method,
        headers,
        credentials: "include",
    });

    const contentType = response.headers.get("content-type") || "";
    const data = contentType.includes("application/json") ? await response.json() : await response.text();

    if (!response.ok) {
        const message = typeof data === "object" && data && data.message
            ? data.message
            : typeof data === "string"
                ? data
                : `Request failed with ${response.status}`;
        throw new Error(message);
    }

    return data;
}

const formValues = (form) => Object.fromEntries(new FormData(form).entries());

async function loadMeta() {
    try {
        state.meta = await apiFetch("/meta");
    } catch {
        state.meta = { version: "Unavailable", commit_sha: "" };
    }
}

async function loadSession() {
    try {
        const data = await apiFetch("/auth/csrf-token");
        state.csrfToken = data.csrfToken || "";
        state.signedIn = true;
        return true;
    } catch {
        state.csrfToken = "";
        state.signedIn = false;
        state.account = null;
        state.notes = [];
        state.accounts = [];
        return false;
    }
}

async function loadAccount() {
    state.account = await apiFetch("/accounts/me");
    return state.account;
}

async function loadNotes() {
    const notes = await apiFetch("/notes");
    state.notes = Array.isArray(notes) ? notes : [];
    return state.notes;
}

async function loadAccounts() {
    const accounts = await apiFetch("/moderator/accounts");
    state.accounts = Array.isArray(accounts) ? accounts : [];
    return state.accounts;
}

async function prepareSignIn() {
    const data = await apiFetch("/auth/pre-session", { method: "POST" });
    state.csrfToken = data.csrfToken || "";
}

async function signUp(form) {
    await apiFetch("/auth/sign-up", {
        method: "POST",
        body: JSON.stringify(formValues(form)),
    });
    state.flash = { message: "Account created. Sign in to continue.", kind: "info" };
    navigate("/sign-in", true);
    render();
}

async function signIn(form) {
    if (!state.csrfToken) {
        await prepareSignIn();
    }

    const data = await apiFetch("/auth/sign-in", {
        method: "POST",
        body: JSON.stringify(formValues(form)),
    });

    state.csrfToken = data.csrfToken || state.csrfToken;
    state.signedIn = true;
    clearFlash();
    await loadMeta();
    await loadAccount();
    await loadNotes();
    navigate("/notes", true);
    render();
}

async function signOut(path) {
    await apiFetch(path, { method: "POST" });
    state.signedIn = false;
    state.csrfToken = "";
    state.account = null;
    state.notes = [];
    state.accounts = [];
    navigate("/sign-in", true);
    render();
}

async function updateUsername(form) {
    await apiFetch("/accounts/me/username", {
        method: "PATCH",
        body: JSON.stringify(formValues(form)),
    });
    await loadAccount();
    setFlash("Username updated.", "info");
}

async function updatePassword(form) {
    await apiFetch("/accounts/me/password", {
        method: "PATCH",
        body: JSON.stringify(formValues(form)),
    });
    setFlash("Password updated.", "info");
}

async function deleteAccount() {
    await apiFetch("/accounts/me", { method: "DELETE" });
    state.signedIn = false;
    state.csrfToken = "";
    state.account = null;
    state.notes = [];
    state.accounts = [];
    navigate("/sign-in", true);
    render();
}

async function createNote() {
    await apiFetch("/notes", { method: "POST" });
    await loadNotes();
    render();
}

async function saveNote(id, body) {
    await apiFetch(`/notes/${id}`, {
        method: "PUT",
        body: JSON.stringify({ body }),
    });
    await loadNotes();
    render();
}

async function removeNote(id) {
    await apiFetch(`/notes/${id}`, { method: "DELETE" });
    await loadNotes();
    render();
}

async function updateAccountRole(id, role) {
    await apiFetch(`/owner/accounts/${id}/role`, {
        method: "PATCH",
        body: JSON.stringify({ role }),
    });
    await loadAccounts();
    setFlash(`Role updated to ${role}.`, "info");
}

async function removeAccount(id) {
    await apiFetch(`/moderator/accounts/${id}`, { method: "DELETE" });
    await loadAccounts();
    render();
}

async function ensureRoute() {
    const path = currentPath();

    if (!state.signedIn) {
        if (path !== "/sign-in" && path !== "/sign-up") {
            navigate("/sign-in", true);
        } else {
            state.route = path;
        }
        return;
    }

    if (path === "/" || path === "/sign-in" || path === "/sign-up") {
        navigate("/notes", true);
        return;
    }

    const knownRoutes = ["/notes", "/account"];
    if (isPrivileged(state.account)) {
        knownRoutes.push("/users");
    }

    if (!knownRoutes.includes(path)) {
        navigate("/notes", true);
    }

    state.route = currentPath();

    if (state.route === "/account") {
        await loadAccount();
    } else if (state.route === "/notes") {
        await loadNotes();
    } else if (state.route === "/users") {
        await loadAccounts();
    }
}

function render() {
    clearNode(root);

    if (!state.signedIn) {
        root.append(authPage({
            mode: state.route === "/sign-up" ? "sign-up" : "sign-in",
            flash: state.flash,
            metaText: state.meta ? `${state.meta.version}${state.meta.commit_sha ? ` · ${state.meta.commit_sha}` : ""}` : "Loading...",
            onNavigateAuth: async (path) => {
                clearFlash();
                navigate(path);
                await ensureRoute();
                render();
            },
            onSignIn: async (form) => {
                try {
                    await signIn(form);
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onSignUp: async (form) => {
                try {
                    await signUp(form);
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
        }));
        return;
    }

    const pages = {
        "/notes": () => notesPage({
            notes: state.notes,
            onCreateNote: async () => {
                try {
                    await createNote();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onRefreshNotes: async () => {
                try {
                    await loadNotes();
                    render();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onSaveNote: async (id, body) => {
                try {
                    await saveNote(id, body);
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onDeleteNote: async (id) => {
                try {
                    await removeNote(id);
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
        }),
        "/account": () => accountPage({
            account: state.account,
            onUpdateUsername: async (form) => {
                try {
                    await updateUsername(form);
                    render();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onUpdatePassword: async (form) => {
                try {
                    await updatePassword(form);
                    render();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onDeleteAccount: async () => {
                try {
                    await deleteAccount();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
        }),
        "/users": () => usersPage({
            accounts: state.accounts,
            actor: state.account,
            roles,
            canUpdateRoles: state.account?.role === "owner",
            onUpdateRole: async (id, role) => {
                try {
                    await updateAccountRole(id, role);
                    render();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onDeleteAccount: async (id) => {
                try {
                    await removeAccount(id);
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
            onRefreshUsers: async () => {
                try {
                    await loadAccounts();
                    render();
                } catch (error) {
                    setFlash(error.message, "error");
                }
            },
        }),
    };

    const route = pages[state.route] ? state.route : "/notes";

    root.append(shell({
        route,
        showUsers: isPrivileged(state.account),
        metaText: state.meta ? `${state.meta.version}${state.meta.commit_sha ? ` · ${state.meta.commit_sha}` : ""}` : "Loading...",
        onNavigate: async (path) => {
            navigate(path);
            await ensureRoute();
            render();
        },
        onSignOut: async () => {
            try {
                await signOut("/auth/sign-out");
            } catch (error) {
                setFlash(error.message, "error");
            }
        },
        content: pages[route](),
    }));
}

async function boot() {
    await loadMeta();
    await loadSession();

    if (state.signedIn) {
        // The account role decides which pages and routes are available
        await loadAccount();
    }

    await ensureRoute();

    if (state.signedIn && state.route === "/notes") {
        await loadNotes();
    }

    render();
}

window.addEventListener("popstate", async () => {
    state.route = currentPath();
    await ensureRoute();
    render();
});

boot().catch((error) => {
    state.flash = { message: error.message, kind: "error" };
    render();
});
