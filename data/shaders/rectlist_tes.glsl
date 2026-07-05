#version 450
layout(quads, equal_spacing, cw) in;

layout(location = 0) in vec4 param_in_0[];
layout(location = 1) in vec4 param_in_1[];
layout(location = 2) in vec4 param_in_2[];
layout(location = 3) in vec4 param_in_3[];
layout(location = 4) in vec4 param_in_4[];
layout(location = 5) in vec4 param_in_5[];
layout(location = 6) in vec4 param_in_6[];
layout(location = 7) in vec4 param_in_7[];
layout(location = 8) in vec4 param_in_8[];
layout(location = 9) in vec4 param_in_9[];
layout(location = 10) in vec4 param_in_10[];
layout(location = 11) in vec4 param_in_11[];
layout(location = 12) in vec4 param_in_12[];
layout(location = 13) in vec4 param_in_13[];
layout(location = 14) in vec4 param_in_14[];
layout(location = 15) in vec4 param_in_15[];

layout(location = 0) out vec4 param_out_0;
layout(location = 1) out vec4 param_out_1;
layout(location = 2) out vec4 param_out_2;
layout(location = 3) out vec4 param_out_3;
layout(location = 4) out vec4 param_out_4;
layout(location = 5) out vec4 param_out_5;
layout(location = 6) out vec4 param_out_6;
layout(location = 7) out vec4 param_out_7;
layout(location = 8) out vec4 param_out_8;
layout(location = 9) out vec4 param_out_9;
layout(location = 10) out vec4 param_out_10;
layout(location = 11) out vec4 param_out_11;
layout(location = 12) out vec4 param_out_12;
layout(location = 13) out vec4 param_out_13;
layout(location = 14) out vec4 param_out_14;
layout(location = 15) out vec4 param_out_15;

void main() {
    int index = int(gl_TessCoord.y) * 2 + int(gl_TessCoord.x);
    gl_Position = gl_in[index].gl_Position;
    param_out_0 = param_in_0[index];
    param_out_1 = param_in_1[index];
    param_out_2 = param_in_2[index];
    param_out_3 = param_in_3[index];
    param_out_4 = param_in_4[index];
    param_out_5 = param_in_5[index];
    param_out_6 = param_in_6[index];
    param_out_7 = param_in_7[index];
    param_out_8 = param_in_8[index];
    param_out_9 = param_in_9[index];
    param_out_10 = param_in_10[index];
    param_out_11 = param_in_11[index];
    param_out_12 = param_in_12[index];
    param_out_13 = param_in_13[index];
    param_out_14 = param_in_14[index];
    param_out_15 = param_in_15[index];
}