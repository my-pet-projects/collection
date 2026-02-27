// Clerk authentication logout script
window.addEventListener("load", async function () {
  let signedOut = false;
  try {
    await Clerk.load();
    if (Clerk.session) {
      await Clerk.signOut();
    }
    signedOut = true;
  } catch (err) {
    console.error("Failed to sign out from Clerk", err);
    const errorEl = document.getElementById("clerk-error");
    if (errorEl) {
      errorEl.textContent =
        "Authentication service unavailable. Please refresh the page.";
      errorEl.classList.remove("hidden");
    }
  }
  if (signedOut) {
    // Redirect to home after signing out
    window.location.replace("/");
  } else {
    // Force reload to attempt recovery
    window.location.reload();
  }
});
