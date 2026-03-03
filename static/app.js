async function nextEpisode(id) {
    try {
        const response = await fetch("/update?id=" + id, {
            method: "POST"
        });

        if (!response.ok) {
            console.error("Update failed");
            return;
        }

        location.reload();

    } catch (error) {
        console.error("Fetch error:", error);
    }
}

async function prevEpisode(id) {
    try {
        const response = await fetch("/decrement?id=" + id, {
            method: "POST"
        });

        if (!response.ok) {
            console.error("Decrement failed");
            return;
        }

        location.reload();

    } catch (error) {
        console.error("Fetch error:", error);
    }
}

async function deleteSeries(id) {
    if (!confirm("Are you sure you want to delete this series?")) {
        return;
    }

    try {
        const response = await fetch("/delete?id=" + id, {
            method: "DELETE"
        });

        if (!response.ok) {
            console.error("Delete failed");
            return;
        }

        location.reload();

    } catch (error) {
        console.error("Fetch error:", error);
    }
}