#version 450
layout(triangles) in;
layout(triangle_strip, max_vertices = 4) out;

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

void emit_vertex(int index) {
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
    EmitVertex();
}

void main() {
    emit_vertex(0);
    emit_vertex(1);
    emit_vertex(2);
    
    gl_Position = gl_in[1].gl_Position + gl_in[2].gl_Position - gl_in[0].gl_Position;
    param_out_0 = param_in_0[1] + param_in_0[2] - param_in_0[0];
    param_out_1 = param_in_1[1] + param_in_1[2] - param_in_1[0];
    param_out_2 = param_in_2[1] + param_in_2[2] - param_in_2[0];
    param_out_3 = param_in_3[1] + param_in_3[2] - param_in_3[0];
    param_out_4 = param_in_4[1] + param_in_4[2] - param_in_4[0];
    param_out_5 = param_in_5[1] + param_in_5[2] - param_in_5[0];
    param_out_6 = param_in_6[1] + param_in_6[2] - param_in_6[0];
    param_out_7 = param_in_7[1] + param_in_7[2] - param_in_7[0];
    param_out_8 = param_in_8[1] + param_in_8[2] - param_in_8[0];
    param_out_9 = param_in_9[1] + param_in_9[2] - param_in_9[0];
    param_out_10 = param_in_10[1] + param_in_10[2] - param_in_10[0];
    param_out_11 = param_in_11[1] + param_in_11[2] - param_in_11[0];
    param_out_12 = param_in_12[1] + param_in_12[2] - param_in_12[0];
    param_out_13 = param_in_13[1] + param_in_13[2] - param_in_13[0];
    param_out_14 = param_in_14[1] + param_in_14[2] - param_in_14[0];
    param_out_15 = param_in_15[1] + param_in_15[2] - param_in_15[0];
    EmitVertex();
    
    EndPrimitive();
}