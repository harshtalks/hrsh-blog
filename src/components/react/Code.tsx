import * as React from "react";
import { Sandpack } from "@codesandbox/sandpack-react";
import { sandpackDark } from "@codesandbox/sandpack-themes";

export const CodeComponent = ({
    files,
    openFile
}: {
    files: any;
    openFile: string;
}) => {
    return (
        <Sandpack
            theme={sandpackDark}
            template="react-ts"
            options={{
                showInlineErrors: true,
                showNavigator: true,
                showLineNumbers: true,
                editorHeight: "400px",
                activeFile: openFile
            }}
            files={files}
        />
    );
};
