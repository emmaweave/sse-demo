import React, { useEffect, useState } from "react";

const ProgressComponent = () => {
  const [progress, setProgress] = useState(0);
  const [isCompleted, setIsCompleted] = useState(false);
  const [projectID, setProjectID] = useState(null);

  let env = "http://localhost:8080";
   env = "https://sse-backend-6d237bcbc6f3.herokuapp.com/";


  const startGeneration = async () => {
    try {
      const response = await fetch(`${env}/generate-assets`);
      const data = await response.json();
      setProjectID(data.projectID);
      setProgress(0);
      setIsCompleted(false); 
    } catch (error) {
      console.error("Error starting generation:", error);
    }
  };

  const downloadAssets = async () => {

  };

  useEffect(() => {
    if (!projectID) return;

    const eventSource = new EventSource(
      `${env}/progress-assets?projectID=${projectID}`
    );

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.progress >= 100) {
          setProgress(100);
          setIsCompleted(true);
          eventSource.close(); 
        } else {
          setProgress(data.progress);
        }
      } catch (error) {
        console.error("Error parsing SSE data:", error);
      }
    };

    eventSource.onerror = () => {
      console.error("SSE connection error");
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [projectID, env]);

  return (
    <div>
      <h1>Asset Generation Progress</h1>
      <button onClick={startGeneration} disabled={progress > 0 && !isCompleted}>
        {progress > 0 && !isCompleted ? "Generating..." : "Start Generation"}
      </button>

      {projectID && (
        <div>
          <label htmlFor="progress-bar">Progress:</label>
          <progress id="progress-bar" value={progress} max="100" />
          <span>{isCompleted ? "Completed" : `${progress}%`}</span>

          {isCompleted && (
            <div>
              <button onClick={downloadAssets}>DOWNLOAD ASSETS</button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ProgressComponent;
