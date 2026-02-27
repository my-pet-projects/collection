// Clerk authentication logout script
window.addEventListener("load", async function () {
  try {
    await Clerk.load();
    if (Clerk.session) {
      await Clerk.signOut();
    }
  } catch (err) {
    console.error("Failed to sign out from Clerk", err);
  }
  // Redirect to home after signing out
  window.location.replace("/");
});
