import * as React from "react";
import {
    Sandpack,
    SandpackCodeEditor,
    SandpackConsole,
    SandpackFileExplorer,
    SandpackLayout,
    SandpackPreview,
    SandpackProvider
} from "@codesandbox/sandpack-react";
import { sandpackDark } from "@codesandbox/sandpack-themes";

export const CodeComponent = () => {
    return (
        <SandpackProvider>
            <Sandpack
                theme={sandpackDark}
                template="vanilla-ts"
                options={{
                    showConsoleButton: true,
                    showInlineErrors: true,
                    showNavigator: true,
                    showLineNumbers: true,
                    showTabs: true,
                    editorHeight: "400px",
                    activeFile: "/index.html",
                    visibleFiles: ["/index.html", "/index.ts"]
                }}
            />
        </SandpackProvider>
    );
};
