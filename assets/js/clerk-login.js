// Clerk authentication initialization for login page
window.__clerkReady = false;

window.addEventListener("load", async function () {
  try {
    await Clerk.load();
    if (Clerk.session) {
      window.location.replace("/workspace/beer");
      return;
    }
    window.__clerkReady = true;
  } catch (err) {
    console.error("Failed to load Clerk", err);
  }
});
