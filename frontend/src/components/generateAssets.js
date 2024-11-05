import React, { useEffect, useState } from 'react';

const ProgressComponent = () => {
    const [progress, setProgress] = useState(0);
    const [isCompleted, setIsCompleted] = useState(false);
    const [assetID, setAssetID] = useState(null);

    const env = "https://sse-demo123-93594a7d6504.herokuapp.com/"

    // Function to start asset generation
    const startGeneration = async () => {
        try {
            const response = await fetch(`${env}/generate`);
            const data = await response.json();
            setAssetID(data.assetID); // Set the assetID received from the backend
            setProgress(0);
            setIsCompleted(false);
        } catch (error) {
            console.error('Error starting generation:', error);
        }
    };

    useEffect(() => {
        if (!assetID) return;

        // Connect to the SSE endpoint with the generated assetID
        const eventSource = new EventSource(`${env}/sse?assetID=${assetID}`);

        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);

            if (data.progress === 'completed') {
                setProgress(100);
                setIsCompleted(true);
                eventSource.close(); // Close the connection when completed
            } else {
                setProgress(data.progress);
            }
        };

        eventSource.onerror = () => {
            console.error('SSE connection error');
            eventSource.close();
        };

        return () => {
            eventSource.close();
        };
    }, [assetID]);

    return (
        <div>
            <h1>Asset Generation Progress</h1>
            <button onClick={startGeneration} disabled={progress > 0 && !isCompleted}>
                {progress > 0 && !isCompleted ? 'Generating...' : 'Start Generation'}
            </button>

            {assetID && (
                <div>
                    <label htmlFor="progress-bar">Progress:</label>
                    <progress id="progress-bar" value={progress} max="100" />
                    <span>{isCompleted ? 'Completed' : `${progress}%`}</span>
                </div>
            )}
        </div>
    );
};

export default ProgressComponent;
