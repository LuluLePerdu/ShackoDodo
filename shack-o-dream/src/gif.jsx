import React, {useEffect, useRef, useState} from "react";

export default function TenorGifButton({ open }) {
    const gifContainer = useRef(null);

    useEffect(() => {
        if (open && gifContainer.current) {
            const script = document.createElement("script");
            script.src = "https://tenor.com/embed.js";
            script.async = true;
            gifContainer.current.appendChild(script);
        }
    }, [open]);

    if (!open) return null;

    return (
        <div
            ref={gifContainer}
            className="tenor-gif-embed"
            data-postid="22954713"
            data-share-method="host"
            data-aspect-ratio="1"
            data-width="100%"
        >
            <a href="https://tenor.com/view/rickroll-roll-rick-never-gonna-give-you-up-never-gonna-gif-22954713">
                Rickroll Never Gonna Give You Up GIF
            </a>{" "}
            from{" "}
            <a href="https://tenor.com/search/rickroll-gifs">Rickroll GIFs</a>
        </div>
    );
}

export function FleeingCloseButton({ open, onClose, delay = 1000 }) {
    const [position, setPosition] = useState({ top: 50, left: 50 });
    const [windowSize, setWindowSize] = useState({ width: window.innerWidth, height: window.innerHeight });
    const [active, setActive] = useState(false);

    useEffect(() => {
        const handleResize = () => setWindowSize({ width: window.innerWidth, height: window.innerHeight });
        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);

    useEffect(() => {
        if (open) {
            const timer = setTimeout(() => setActive(true), delay);
            return () => clearTimeout(timer);
        } else {
            setActive(false);
        }
    }, [open, delay]);

    const handleMouseEnter = () => {
        if (!active) return;
        const top = Math.floor(Math.random() * (windowSize.height - 50)); // 50px hauteur du bouton
        const left = Math.floor(Math.random() * (windowSize.width - 100)); // 100px largeur du bouton
        setPosition({ top, left });
    };

    if (!open) return null;

    return (
        <button
            onMouseEnter={handleMouseEnter}
            onClick={onClose}
            style={{
                position: "absolute",
                top: position.top,
                left: position.left,
                transition: "top 1.5s, left 1.5s",
                zIndex: 100,
                padding: "10px 20px",
                cursor: "pointer",
            }}
        >
            Fermer GIF
        </button>
    );
}
