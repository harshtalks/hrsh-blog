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
        <div>
            <Sandpack
                theme={sandpackDark}
                template="astro"
                options={{
                    showConsoleButton: true,
                    showInlineErrors: true,
                    showNavigator: true,
                    showLineNumbers: true,
                    showTabs: true,
                    editorHeight: "400px"
                }}
            />
        </div>
    );
};
