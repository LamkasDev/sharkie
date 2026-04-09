#version 450

layout(push_constant) uniform PC {
    float time;
} pc;
layout(location = 0) out vec2 fragPos;

vec2 positions[3] = vec2[](
    vec2(0.0, -0.5),
    vec2(0.5, 0.5),
    vec2(-0.5, 0.5)
);

void main() {
    vec2 pos = positions[gl_VertexIndex];
    fragPos = pos;

    gl_Position = vec4(pos, 0.0, 1.0);
    gl_Position.y = -gl_Position.y;
}