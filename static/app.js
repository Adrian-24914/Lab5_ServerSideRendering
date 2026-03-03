async function nextEpisode(id) {
    try {
        const response = await fetch("/update?id=" + id, {
            method: "POST"
        });

        if (!response.ok) {
            console.error("Update failed");
            return;
        }

        // Recargar la página para ver el cambio
        location.reload();

    } catch (error) {
        console.error("Fetch error:", error);
    }
}