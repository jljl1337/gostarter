import { test, expect } from '@playwright/test';
const password = 'P@ssw0rd!';

// Helper to build full url for API paths
function api(path: string) {
  return `/api${path}`;
}

test.describe('Full example API', () => {
  let preSessionCookie: string | null = null;
  let csrfToken: string | null = null;

  test('e2e test', async ({ request: baseRequest }) => {
    const username = `testuser${Math.floor(Math.random() * 9000) + 1000}`; // 4 digits -> stays within 3-20 chars
    let currentPassword = password;
    // 1) Sign up
    const signUp = await baseRequest.post(api('/auth/sign-up'), {
      headers: { 'Content-Type': 'application/json' },
      data: JSON.stringify({
        username,
        password,
        languageCode: 'en-US',
      }),
    });
    expect(signUp.status()).toBe(201);

    // 2) Pre-session (expect cookie + csrf token)
    const pre = await baseRequest.post(api('/auth/pre-session'));
    expect(pre.status()).toBe(200);
    const preJson = await pre.json();
    expect(preJson.csrfToken).toBeTruthy();
    csrfToken = preJson.csrfToken;

    // collect pre-session cookie
    const setCookieHeader = pre.headers()['set-cookie'];
    expect(setCookieHeader).toBeTruthy();
    // keep only the cookie name=value (strip attributes)
    preSessionCookie = setCookieHeader.split(';')[0];
    expect(preSessionCookie).toContain('session_token');

    // 3) Sign in using pre-session cookie and CSRF header
    const signIn = await baseRequest.post(api('/auth/sign-in'), {
      headers: {
        'X-CSRF-Token': csrfToken || '',
        cookie: preSessionCookie || '',
        'Content-Type': 'application/json',
      },
      data: JSON.stringify({
        username,
        password: currentPassword,
      }),
    });
    expect(signIn.status()).toBe(200);
    const signInJson = await signIn.json();
    expect(signInJson.csrfToken).toBeTruthy();
    // collect active session cookie
    const activeCookie = signIn.headers()['set-cookie'];
    expect(activeCookie).toBeTruthy();
    let activeSessionCookie = Array.isArray(activeCookie) ? activeCookie[0] : activeCookie;
    // keep only the cookie name=value (strip attributes)
    activeSessionCookie = activeSessionCookie.split(';')[0];
    expect(activeSessionCookie).toContain('session_token');

    // Use active session cookie + session CSRF token for subsequent authenticated requests
    const sessionCsrf = signInJson.csrfToken;
    const authHeaders = { cookie: activeSessionCookie, 'X-CSRF-Token': sessionCsrf };

    // 4) Get current account
    const me = await baseRequest.get(api('/accounts/me'), { headers: authHeaders });
    expect(me.status()).toBe(200);
    const meJson = await me.json();
    expect(meJson.username).toBe(username);

    // 5) Update username
    const newUsername = username + '_2';
    const updUser = await baseRequest.patch(api('/accounts/me/username'), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ newUsername }),
    });
    expect(updUser.status()).toBe(200);

    // verify updated
    const me2 = await baseRequest.get(api('/accounts/me'), { headers: authHeaders });
    expect(me2.status()).toBe(200);
    const me2Json = await me2.json();
    expect(me2Json.username).toBe(newUsername);

    // 6) Update password
    const updPass = await baseRequest.patch(api('/accounts/me/password'), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ oldPassword: password, newPassword: 'N3wP@ssw0rd!' }),
    });
    expect(updPass.status()).toBe(200);
    currentPassword = 'N3wP@ssw0rd!';

    // 7) Notes endpoint lifecycle
    const createNote = await baseRequest.post(api('/notes'), { headers: authHeaders });
    expect(createNote.status()).toBe(201);

    const listNotes = await baseRequest.get(api('/notes'), { headers: authHeaders });
    expect(listNotes.status()).toBe(200);
    const notes = await listNotes.json();
    expect(Array.isArray(notes)).toBeTruthy();
    expect(notes.length).toBeGreaterThan(0);

    const createdNote = notes[notes.length - 1];
    expect(createdNote).toBeTruthy();
    expect(createdNote.body).toBe('This is a new note.');

    const noteId = createdNote.id;
    const updatedBody = 'Updated note body';
    const updateNote = await baseRequest.put(api(`/notes/${noteId}`), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ body: updatedBody }),
    });
    expect(updateNote.status()).toBe(200);

    const listNotesAfterUpdate = await baseRequest.get(api('/notes'), { headers: authHeaders });
    expect(listNotesAfterUpdate.status()).toBe(200);
    const updatedNotes = await listNotesAfterUpdate.json();
    const updatedNote = updatedNotes.find((note: { id: string }) => note.id === noteId);
    expect(updatedNote).toBeTruthy();
    expect(updatedNote.body).toBe(updatedBody);

    const deleteNote = await baseRequest.delete(api(`/notes/${noteId}`), { headers: authHeaders });
    expect(deleteNote.status()).toBe(200);

    const listNotesAfterDelete = await baseRequest.get(api('/notes'), { headers: authHeaders });
    expect(listNotesAfterDelete.status()).toBe(200);
    const remainingNotes = await listNotesAfterDelete.json();
    expect(remainingNotes.find((note: { id: string }) => note.id === noteId)).toBeUndefined();

    // 8) Get CSRF token endpoint (should succeed)
    const csrf = await baseRequest.get(api('/auth/csrf-token'), { headers: authHeaders });
    expect(csrf.status()).toBe(200);
    const csrfJson = await csrf.json();
    expect(csrfJson.csrfToken).toBeTruthy();

    // 9) Sign out
    const signOut = await baseRequest.post(api('/auth/sign-out'), { headers: authHeaders });
    expect(signOut.status()).toBe(200);

    const signOutCookie = signOut.headers()['set-cookie'];
    expect(signOutCookie).toBeTruthy();
    const expiredSessionCookie = (Array.isArray(signOutCookie) ? signOutCookie[0] : signOutCookie).split(';')[0];
    expect(expiredSessionCookie).toContain('session_token=');

    // 10) Ensure protected endpoint returns unauthorized after sign out
    const meAfterSignOut = await baseRequest.get(api('/accounts/me'), { headers: { cookie: expiredSessionCookie } });
    expect(meAfterSignOut.status()).toBe(401);

    // 11) Create a fresh session for delete-account coverage
    const preAgain = await baseRequest.post(api('/auth/pre-session'));
    expect(preAgain.status()).toBe(200);
    const preAgainJson = await preAgain.json();
    expect(preAgainJson.csrfToken).toBeTruthy();

    const preAgainCookie = preAgain.headers()['set-cookie'];
    expect(preAgainCookie).toBeTruthy();
    const preAgainSessionCookie = (Array.isArray(preAgainCookie) ? preAgainCookie[0] : preAgainCookie).split(';')[0];
    expect(preAgainSessionCookie).toContain('session_token');

    const signInAgain = await baseRequest.post(api('/auth/sign-in'), {
      headers: {
        'X-CSRF-Token': preAgainJson.csrfToken,
        cookie: preAgainSessionCookie,
        'Content-Type': 'application/json',
      },
      data: JSON.stringify({
        username: newUsername,
        password: currentPassword,
      }),
    });
    expect(signInAgain.status()).toBe(200);
    const signInAgainJson = await signInAgain.json();
    expect(signInAgainJson.csrfToken).toBeTruthy();

    const activeAgainCookie = signInAgain.headers()['set-cookie'];
    expect(activeAgainCookie).toBeTruthy();
    const activeAgainSessionCookie = (Array.isArray(activeAgainCookie) ? activeAgainCookie[0] : activeAgainCookie).split(';')[0];
    expect(activeAgainSessionCookie).toContain('session_token');

    const authHeadersAgain = { cookie: activeAgainSessionCookie, 'X-CSRF-Token': signInAgainJson.csrfToken };

    // 12) Delete account
    const deleteAccount = await baseRequest.delete(api('/accounts/me'), { headers: authHeadersAgain });
    expect(deleteAccount.status()).toBe(200);

    const deleteCookie = deleteAccount.headers()['set-cookie'];
    expect(deleteCookie).toBeTruthy();
    expect(deleteCookie).toContain('session_token=');

    // 13) Ensure protected endpoint returns unauthorized after account deletion
    const meAfter = await baseRequest.get(api('/accounts/me'), { headers: { cookie: deleteCookie } });
    expect(meAfter.status()).toBe(401);
  });
});
