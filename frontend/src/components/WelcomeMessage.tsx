import { useState, useEffect } from "react";

const WelcomeMessage = () => {
  const [message, setMessage] = useState("");

  useEffect(() => {
    fetch("http://localhost:13013/")
      .then((response) => response.json())
      .then((data) => {
        console.log("Received data:", data);
        setMessage(data.message);
      })
      .catch((error) => {
        console.error("Error:", error);
        setMessage("Error al cargar el mensaje");
      });
  }, []);

  return (
    <div className="text-center mt-10">
      <h1 className="text-2xl font-bold">{message || "Cargando..."}</h1>
    </div>
  );
};

export default WelcomeMessage;
