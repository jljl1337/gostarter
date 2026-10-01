import { test, expect, type APIRequestContext } from '@playwright/test';
const password = 'P@ssw0rd!';

type Auth = { cookie: string; csrf: string };
type Account = { id: string; username: string; role: string; languageCode: string; createdAt: string };

// Helper to build full url for API paths
function api(path: string) {
  return `/api${path}`;
}

// Helper to keep only the cookie name=value (strip attributes)
function sessionCookieOf(response: { headers: () => Record<string, string | string[]> }) {
  const setCookie = response.headers()['set-cookie'];
  return (Array.isArray(setCookie) ? setCookie[0] : setCookie).split(';')[0];
}

function authOf(auth: Auth) {
  return { cookie: auth.cookie, 'X-CSRF-Token': auth.csrf };
}

// Helper to sign up an account and sign it in, returns its session headers
async function signUpAndSignIn(request: APIRequestContext, username: string, accountPassword: string): Promise<Auth> {
  const signUp = await request.post(api('/auth/sign-up'), {
    headers: { 'Content-Type': 'application/json' },
    data: JSON.stringify({
      username,
      password: accountPassword,
      languageCode: 'en-US',
    }),
  });
  expect(signUp.status()).toBe(201);

  const preSession = await request.post(api('/auth/pre-session'));
  expect(preSession.status()).toBe(200);
  const preSessionJson = await preSession.json();
  expect(preSessionJson.csrfToken).toBeTruthy();
  const preSessionCookie = sessionCookieOf(preSession);
  expect(preSessionCookie).toContain('session_token');

  const signIn = await request.post(api('/auth/sign-in'), {
    headers: {
      'X-CSRF-Token': preSessionJson.csrfToken,
      cookie: preSessionCookie,
      'Content-Type': 'application/json',
    },
    data: JSON.stringify({
      username,
      password: accountPassword,
    }),
  });
  expect(signIn.status()).toBe(200);
  const signInJson = await signIn.json();
  expect(signInJson.csrfToken).toBeTruthy();

  const sessionCookie = sessionCookieOf(signIn);
  expect(sessionCookie).toContain('session_token');

  return { cookie: sessionCookie, csrf: signInJson.csrfToken };
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

    // Pre-session (expect cookie + csrf token)
    const pre = await baseRequest.post(api('/auth/pre-session'));
    expect(pre.status()).toBe(200);
    const preJson = await pre.json();
    expect(preJson.csrfToken).toBeTruthy();
    csrfToken = preJson.csrfToken;

    // collect pre-session cookie
    const setCookieHeader = pre.headers()['set-cookie'];
    expect(setCookieHeader).toBeTruthy();
    // keep only the cookie name=value (strip attributes)
    preSessionCookie = sessionCookieOf(pre);
    expect(preSessionCookie).toContain('session_token');

    // Sign in using pre-session cookie and CSRF header
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
    // keep only the cookie name=value (strip attributes)
    const activeSessionCookie = sessionCookieOf(signIn);
    expect(activeSessionCookie).toContain('session_token');

    // Use active session cookie + session CSRF token for subsequent authenticated requests
    const sessionCsrf = signInJson.csrfToken;
    const authHeaders = { cookie: activeSessionCookie, 'X-CSRF-Token': sessionCsrf };

    // Get current account
    const me = await baseRequest.get(api('/accounts/me'), { headers: authHeaders });
    expect(me.status()).toBe(200);
    const meJson = await me.json();
    expect(meJson.username).toBe(username);

    // The first account of a fresh database is the owner
    expect(meJson.role).toBe('owner');
    const ownerAccountID = meJson.id;

    // Update username
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

    // Update language code
    const updLang = await baseRequest.patch(api('/accounts/me/language'), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ languageCode: 'fr-FR' }),
    });
    expect(updLang.status()).toBe(200);

    // verify updated
    const me3 = await baseRequest.get(api('/accounts/me'), { headers: authHeaders });
    expect(me3.status()).toBe(200);
    const me3Json = await me3.json();
    expect(me3Json.languageCode).toBe('fr-FR');

    // Update password
    const updPass = await baseRequest.patch(api('/accounts/me/password'), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ oldPassword: password, newPassword: 'N3wP@ssw0rd!' }),
    });
    expect(updPass.status()).toBe(200);
    currentPassword = 'N3wP@ssw0rd!';

    // Role gated endpoints
    // Every account but the first one is created with the "user" role
    const moderatorUsername = `moduser${Math.floor(Math.random() * 9000) + 1000}`;
    const memberUsername = `member${Math.floor(Math.random() * 9000) + 1000}`;

    const moderatorAuth = await signUpAndSignIn(baseRequest, moderatorUsername, password);
    const memberAuth = await signUpAndSignIn(baseRequest, memberUsername, password);

    const moderatorMe = await baseRequest.get(api('/accounts/me'), { headers: authOf(moderatorAuth) });
    expect(moderatorMe.status()).toBe(200);
    const moderatorJson = await moderatorMe.json();
    expect(moderatorJson.role).toBe('user');

    const memberMe = await baseRequest.get(api('/accounts/me'), { headers: authOf(memberAuth) });
    expect(memberMe.status()).toBe(200);
    const memberJson = await memberMe.json();
    expect(memberJson.role).toBe('user');

    // A user cannot reach the role gated routes, the server answers as if the
    // route did not exist
    const userList = await baseRequest.get(api('/moderator/accounts'), { headers: authOf(moderatorAuth) });
    expect(userList.status()).toBe(404);

    const userPromote = await baseRequest.patch(api(`/owner/accounts/${memberJson.id}/role`), {
      headers: { ...authOf(moderatorAuth), 'Content-Type': 'application/json' },
      data: JSON.stringify({ role: 'moderator' }),
    });
    expect(userPromote.status()).toBe(404);

    // An owner can list the accounts of every role
    const ownerList = await baseRequest.get(api('/moderator/accounts'), { headers: authHeaders });
    expect(ownerList.status()).toBe(200);
    const accounts: Account[] = await ownerList.json();
    expect(Array.isArray(accounts)).toBeTruthy();
    expect(accounts.length).toBe(3);
    expect(JSON.stringify(accounts)).not.toContain('passwordHash');
    expect(accounts.find((account) => account.id === ownerAccountID)?.role).toBe('owner');
    expect(accounts.find((account) => account.id === moderatorJson.id)?.role).toBe('user');

    // An owner can promote a user to moderator
    const promote = await baseRequest.patch(api(`/owner/accounts/${moderatorJson.id}/role`), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ role: 'moderator' }),
    });
    expect(promote.status()).toBe(200);

    const promotedMe = await baseRequest.get(api('/accounts/me'), { headers: authOf(moderatorAuth) });
    expect(promotedMe.status()).toBe(200);
    expect((await promotedMe.json()).role).toBe('moderator');

    // An unknown role is rejected
    const unknownRole = await baseRequest.patch(api(`/owner/accounts/${moderatorJson.id}/role`), {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: JSON.stringify({ role: 'admin' }),
    });
    expect(unknownRole.status()).toBe(422);

    // A moderator cannot promote accounts
    const moderatorPromote = await baseRequest.patch(api(`/owner/accounts/${memberJson.id}/role`), {
      headers: { ...authOf(moderatorAuth), 'Content-Type': 'application/json' },
      data: JSON.stringify({ role: 'moderator' }),
    });
    expect(moderatorPromote.status()).toBe(404);

    // A moderator cannot delete an account of the same or of a higher role
    const moderatorDeleteOwner = await baseRequest.delete(api(`/moderator/accounts/${ownerAccountID}`), {
      headers: authOf(moderatorAuth),
    });
    expect(moderatorDeleteOwner.status()).toBe(403);

    const moderatorDeleteSelf = await baseRequest.delete(api(`/moderator/accounts/${moderatorJson.id}`), {
      headers: authOf(moderatorAuth),
    });
    expect(moderatorDeleteSelf.status()).toBe(403);

    // A moderator can delete a user, which also removes its sessions
    const moderatorDeleteUser = await baseRequest.delete(api(`/moderator/accounts/${memberJson.id}`), {
      headers: authOf(moderatorAuth),
    });
    expect(moderatorDeleteUser.status()).toBe(200);

    const deletedMember = await baseRequest.get(api('/accounts/me'), { headers: authOf(memberAuth) });
    expect(deletedMember.status()).toBe(401);

    const listAfterDelete = await baseRequest.get(api('/moderator/accounts'), { headers: authHeaders });
    expect(listAfterDelete.status()).toBe(200);
    expect((await listAfterDelete.json()).length).toBe(2);

    // Deleting an unknown account is not found
    const deleteUnknown = await baseRequest.delete(api(`/moderator/accounts/${memberJson.id}`), {
      headers: authOf(moderatorAuth),
    });
    expect(deleteUnknown.status()).toBe(404);

    // Notes endpoint lifecycle
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
    const updatedBody = 'good good bad note';
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
    expect(updatedNote.positivity).toBe(0);

    // Test positivity queue
    await new Promise((resolve) => setTimeout(resolve, 1000)); // wait 1 second for sentiment analysis to complete

    const secondListNotes = await baseRequest.get(api('/notes'), { headers: authHeaders });
    expect(secondListNotes.status()).toBe(200);
    const secondNotes = await secondListNotes.json();
    expect(Array.isArray(secondNotes)).toBeTruthy();
    const secondNote = secondNotes.find((note: { id: string }) => note.id === noteId);
    expect(secondNote).toBeTruthy();
    expect(secondNote.body).toBe(updatedBody);
    expect(secondNote.positivity).toBe(1);

    // Continue with note deletion
    const deleteNote = await baseRequest.delete(api(`/notes/${noteId}`), { headers: authHeaders });
    expect(deleteNote.status()).toBe(200);

    const listNotesAfterDelete = await baseRequest.get(api('/notes'), { headers: authHeaders });
    expect(listNotesAfterDelete.status()).toBe(200);
    const remainingNotes = await listNotesAfterDelete.json();
    expect(remainingNotes.find((note: { id: string }) => note.id === noteId)).toBeUndefined();

    // Get CSRF token endpoint (should succeed)
    const csrf = await baseRequest.get(api('/auth/csrf-token'), { headers: authHeaders });
    expect(csrf.status()).toBe(200);
    const csrfJson = await csrf.json();
    expect(csrfJson.csrfToken).toBeTruthy();

    // Sign out
    const signOut = await baseRequest.post(api('/auth/sign-out'), { headers: authHeaders });
    expect(signOut.status()).toBe(200);

    const signOutCookie = signOut.headers()['set-cookie'];
    expect(signOutCookie).toBeTruthy();
    const expiredSessionCookie = sessionCookieOf(signOut);
    expect(expiredSessionCookie).toContain('session_token=');

    // Ensure protected endpoint returns unauthorized after sign out
    const meAfterSignOut = await baseRequest.get(api('/accounts/me'), { headers: { cookie: expiredSessionCookie } });
    expect(meAfterSignOut.status()).toBe(401);

    // Create a fresh session for delete-account coverage
    const preAgain = await baseRequest.post(api('/auth/pre-session'));
    expect(preAgain.status()).toBe(200);
    const preAgainJson = await preAgain.json();
    expect(preAgainJson.csrfToken).toBeTruthy();

    const preAgainCookie = preAgain.headers()['set-cookie'];
    expect(preAgainCookie).toBeTruthy();
    const preAgainSessionCookie = sessionCookieOf(preAgain);
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
    const activeAgainSessionCookie = sessionCookieOf(signInAgain);
    expect(activeAgainSessionCookie).toContain('session_token');

    const authHeadersAgain = { cookie: activeAgainSessionCookie, 'X-CSRF-Token': signInAgainJson.csrfToken };

    // Delete account
    const deleteAccount = await baseRequest.delete(api('/accounts/me'), { headers: authHeadersAgain });
    expect(deleteAccount.status()).toBe(200);

    const deleteCookie = deleteAccount.headers()['set-cookie'];
    expect(deleteCookie).toBeTruthy();
    expect(deleteCookie).toContain('session_token=');

    // Ensure protected endpoint returns unauthorized after account deletion
    const meAfter = await baseRequest.get(api('/accounts/me'), { headers: { cookie: deleteCookie } });
    expect(meAfter.status()).toBe(401);
  });
});
